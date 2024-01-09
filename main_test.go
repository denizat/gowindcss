package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestParseString(t *testing.T) {
	assert := assert.New(t)
	bs := MakeBaseClasses(nil)
	b, err := os.ReadFile("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
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

func FuzzParseString(f *testing.F) {
	f.Add("aspect-video")
	f.Add("-[0-[")
	bs := MakeBaseClasses(nil)
	f.Fuzz(func(t *testing.T, s string) {
		ParseString(s, variants, bs)
	})
}
