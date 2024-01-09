package main

import (
	"testing"
)

func TestParseString(t *testing.T) {
	s := "aspect-video"
	bs := MakeBaseClasses(nil)
	csses := ParseString(s, variants, bs)
	if len(csses) != 1 {
		t.Error("len(csses) != 1")
	}
	sel := csses[0].css.Selector
	if sel != s {
		t.Errorf("sel:%s != %s", sel, s)
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
