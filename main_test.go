package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"regexp"
	"strings"
	"testing"
)

var commentLineRegex = regexp.MustCompile("#.*\n")

type testCase struct {
	from string
	to   string
}

func parseTestFile(fileName string) ([]testCase, error) {
	b, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	s := string(b)
	s = commentLineRegex.ReplaceAllLiteralString(s, "")
	s = strings.TrimSpace(s)
	casesStrings := strings.Split(s, "\n\n")
	cases := []testCase{}
	for _, casesString := range casesStrings {
		parts := strings.SplitN(casesString, "\n", 2)
		cases = append(cases, testCase{from: parts[0], to: parts[1]})
	}
	return cases, nil
}

func Helper(fileName string, t *testing.T, bs map[string]OrderedCSS) {
	assert := assert.New(t)
	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	cases := strings.Split(s, "\n\n")
	for _, v := range cases {
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

func TestFormat(t *testing.T) {
	assert := assert.New(t)
	cases, err := parseTestFile("tests/formattertests.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, acase := range cases {
		r := strings.NewReader(acase.from)
		var sb strings.Builder

		Format(r, &sb, variants, MakeBaseClasses(nil))
		assert.Equal(acase.to, sb.String())
	}
}

func FuzzParseString(f *testing.F) {
	f.Add("aspect-video")
	f.Add("-[0-[")
	bs := MakeBaseClasses(nil)
	f.Fuzz(func(t *testing.T, s string) {
		ParseString(s, variants, bs)
	})
}

func FuzzFormat(f *testing.F) {
	bs := MakeBaseClasses(nil)
	var nb = NilByteWriter{}
	f.Fuzz(func(t *testing.T, s string) {
		r := strings.NewReader(s)
		Format(r, nb, variants, bs)
	})
}
