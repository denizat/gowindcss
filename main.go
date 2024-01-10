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
	"text/tabwriter"
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
	var dump bool
	flag.BoolVar(&dump, "d", false, "dump all css generated from base classes")
	formatFile := flag.String("f", "", "format the file")
	// repl that can print out what class names will become and prints out their order number, but also
	// can do formatting if you start the line with .f
	repl := flag.Bool("r", false, "interactive mode")
	_ = repl
	list := flag.Bool("l", false, "list out base classes in order and the declarations they will generate")
	_ = flag.String("i", "", "input css file")
	_ = flag.String("o", "", "output css file") // or could do redirection?
	flag.Parse()

	var bs map[string]OrderedCSS
	if configFileName != "" {
		bs = HandleConfigFile(&configFileName)
	} else {
		bs = MakeBaseClasses(nil)
	}

	if dump {
		ks := []OrderedCSS{}
		for k, v := range bs {
			v.css.Selector = k
			ks = append(ks, v)
		}
		slices.SortFunc(ks, func(a, b OrderedCSS) int {
			return OrderedCSSLess(a, b)
		})
		fmt.Println(OrderedCSSArrToString(ks))
		os.Exit(0)
	}

	if *list {
		ks := []OrderedCSS{}
		for k, v := range bs {
			v.css.Selector = k
			ks = append(ks, v)
		}
		slices.SortFunc(ks, func(a, b OrderedCSS) int {
			return OrderedCSSLess(a, b)
		})
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "class\torder number\tdeclarations\t")
		for _, c := range ks {
			fmt.Fprintf(w, "%s\t%d\t", c.css.Selector, c.order)
			for _, d := range c.css.Declarations {
				fmt.Fprintf(w, "%s: %s;\t", d.Property, d.Value)
			}
			fmt.Fprintln(w)
		}
		w.Flush()
		os.Exit(0)
	}

	if *formatFile != "" {
		inFile, err := os.Open(*formatFile)
		if err != nil {
			panic(err)
		}
		outFile, err := os.OpenFile(*formatFile, os.O_RDWR, 0777)
		if err != nil {
			panic(err)
		}
		reader := bufio.NewReader(inFile)
		writer := bufio.NewWriter(outFile)
		Format(reader, writer, variants, bs)
		os.Exit(0)
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

func ParseConfigFile(file []byte) map[string]OrderedCSS {
	var config Config
	json.Unmarshal(file, &config)
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

func HandleConfigFile(fileName *string) map[string]OrderedCSS {
	if fileName == nil {
		return MakeBaseClasses(nil)
	}
	bs, err := os.ReadFile(*fileName)
	if err != nil {
		panic(err)
	}
	return ParseConfigFile(bs)
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
		ac := ParseString(a, variants, bs)[0]
		bc := ParseString(b, variants, bs)[0]
		return OrderedCSSLess(ac, bc)
	})
	return strings.Join(classNames, " ")
}
