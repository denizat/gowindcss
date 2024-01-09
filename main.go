package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"os"
	"sort"
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
