package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"sort"
	"strings"
)

var cache map[string]OrderedCSS

func dumpCache() []OrderedCSS {
	arr := []OrderedCSS{}
	for _, v := range cache {
		arr = append(arr, v)
	}
	sort.Slice(arr, func(i, j int) bool {
		// I can figure out more about tailwinds order by
		// dumping my scrape into their playground
		return true
	})
	return arr
}

func main() {
	var configFileName string
	flag.StringVar(&configFileName, "c", "", "location of a configuration file, optional")
	flag.Parse()

	// read config and create base classes

	var bs map[string]OrderedCSS
	if configFileName != "" {
		bs = HandleConfigFile(&configFileName)
	} else {
		bs = MakeBaseClasses(nil)
	}

	// do the rest
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	for scanner.Scan() {
		csses := ParseString(scanner.Text(), variants, bs)
		s := OrderedCSSArrToString(csses)
		_, err := writer.WriteString(s)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
		}
		err = writer.Flush()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
		}
	}
}

func HandleConfigFile(fileName *string) map[string]OrderedCSS {
	if fileName == nil {
		return MakeBaseClasses(nil)
	}
	bs, err := os.ReadFile(*fileName)
	if err != nil {
		panic(err)
	}
	var config Config
	json.Unmarshal(bs, &config)
	if config.Theme.Colors != nil {
		defaultColors = config.Theme.Colors
	}
	if config.Theme.Extend.Colors != nil {
		maps.Copy(defaultColors, config.Theme.Extend.Colors)
	}
	if config.Theme.Screens != nil {
		defaultBreakpoints = config.Theme.Screens
	}
	if config.Theme.Extend.Screens != nil {
		maps.Copy(defaultColors, config.Theme.Extend.Screens)
	}
	return MakeBaseClasses(&config)
}

func Format(r io.ByteReader, w io.ByteWriter, vs map[string]Variant, bs map[string]OrderedCSS) {
	for {
		streamUntilMatch(r, w, "class=\"")
		s, err := collectUntil(r, '"')
		if err != nil {
			break
		}
		s = realFormatGiveBetterName(s, vs, bs) + "\""
		for i := range s {
			w.WriteByte(s[i])
		}
	}
}

func collectUntil(r io.ByteReader, c byte) (string, error) {
	var sb strings.Builder
	for {
		b, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		if b == c {
			return sb.String(), nil
		}
		sb.WriteByte(b)
	}
}

func streamUntilMatch(r io.ByteReader, w io.ByteWriter, s string) {
	i := 0
	for {
		b, err := r.ReadByte()
		if err != nil {
			return
		}
		// TODO: handle this error better
		if w.WriteByte(b) != nil {
			return
		}
		if b == s[i] {
			i++
			if i == len(s) {
				return
			}
		} else if i > 0 {
			i = 0
		}
	}
}

func realFormatGiveBetterName(s string, vs map[string]Variant, bs map[string]OrderedCSS) string {
	classNames := strings.Split(s, " ")
	slices.SortFunc(classNames, func(a, b string) int {
		return OrderedCSSLess(a, b, vs, bs)
	})
	return strings.Join(classNames, " ")
}
