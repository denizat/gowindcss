package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"regexp"
	"strings"
	"testing"
)

var commentLineRegex = regexp.MustCompile("#.*\n")

func Helper(fileName string, t *testing.T, bs map[string]OrderedCSS) {
	assert := assert.New(t)
	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	s = commentLineRegex.ReplaceAllLiteralString(s, "")
	cases := strings.Split(s, "\n\n")
	for _, v := range cases[len(cases)-1:] {
		parts := strings.SplitN(v, "\n", 2)
		className := parts[0]
		target := parts[1]
		cs := ParseString(className, variants, bs)
		res := OrderedCSSArrToString(cs)
		res = strings.TrimSpace(res)
		target = strings.TrimSpace(target)
		assert.Equal(target, res)
	}
}

func TestParseString(t *testing.T) {
	bs := HandleConfigFile(nil)
	Helper("tests/defaultstests.txt", t, bs)
	//fileName := "tests/config.json"
	//bs = HandleConfigFile(&fileName)
	//Helper("tests/configtests.txt", t, bs)
}

func FuzzParseString(f *testing.F) {
	f.Add("aspect-video")
	f.Add("-[0-[")
	bs := MakeBaseClasses(nil)
	f.Fuzz(func(t *testing.T, s string) {
		ParseString(s, variants, bs)
	})
}
