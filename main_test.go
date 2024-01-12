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

func TestParse(t *testing.T) {
	assert := assert.New(t)
	var s string
	var ex fullClassInformation
	var act fullClassInformation

	s = "abc"
	ex = fullClassInformation{class: parsedValue{name: "abc", arbitraryText: ""}}
	act = *parse(s)
	assert.Equal(ex, act)

	s = "abc-abc"
	ex = fullClassInformation{class: parsedValue{name: "abc-abc", arbitraryText: ""}}
	act = *parse(s)
	assert.Equal(ex, act)

	s = "abc-abc-[100]"
	ex = fullClassInformation{class: parsedValue{name: "abc-abc", arbitraryText: "100"}}
	act = *parse(s)
	assert.Equal(ex, act)

	s = "a:b"
	ex = fullClassInformation{
		variants: []parsedValue{
			{name: "a", arbitraryText: ""},
		},
		class: parsedValue{name: "b", arbitraryText: ""}}
	act = *parse(s)
	assert.Equal(ex, act)

	s = "a-[100]:b-[200]"
	ex = fullClassInformation{
		variants: []parsedValue{
			{name: "a", arbitraryText: "100"},
		},
		class: parsedValue{name: "b", arbitraryText: "200"}}
	act = *parse(s)
	assert.Equal(ex, act)

	s = "[100]:[200]:[300]"
	ex = fullClassInformation{
		variants: []parsedValue{
			{name: "", arbitraryText: "100"},
			{name: "", arbitraryText: "200"},
		},
		class: parsedValue{name: "", arbitraryText: "300"}}
	act = *parse(s)
	assert.Equal(ex, act)

	s = "a/b"
	ex = fullClassInformation{
		class: parsedValue{name: "a", slashText: "b"}}
	act = *parse(s)
	assert.Equal(ex, act)

	s = "a-[zz]/b:c"
	ex = fullClassInformation{
		variants: []parsedValue{{name: "a", arbitraryText: "zz", slashText: "b"}},
		class:    parsedValue{name: "c"},
	}
	act = *parse(s)
	assert.Equal(ex, act)

	s = "a-[zz]/b-[car]:c/d"
	ex = fullClassInformation{
		variants: []parsedValue{{name: "a", arbitraryText: "zz", slashText: "b-[car]"}},
		class:    parsedValue{name: "c", slashText: "d"},
	}
	act = *parse(s)
	assert.Equal(ex, act)

}

func FuzzParseString(f *testing.F) {
	f.Add("aspect-video")
	f.Add("-[0-[")
	bs := MakeBaseClasses(nil)
	f.Fuzz(func(t *testing.T, s string) {
		ParseString(s, variants, bs)
	})
}

type NilByteWriter struct{}

func (n NilByteWriter) WriteByte(_ byte) error {
	return nil
}

func FuzzFormat(f *testing.F) {

	bs := MakeBaseClasses(nil)
	var nb = NilByteWriter{}
	f.Fuzz(func(t *testing.T, s string) {
		r := strings.NewReader(s)
		Format(r, nb, variants, bs)
	})
}
