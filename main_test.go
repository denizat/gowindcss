package main

import (
	"testing"
)

func TestThing(t *testing.T) {
	csses := ParseString("aspect-[")
	s := OrderedCSSArrToString(csses)
	t.Logf(s)
}

func TestDefaultColors(t *testing.T) {
	m := map[string]string{}
	for k, v := range defaultColors {
		for kk, vv := range v {
			//m[k+"-"+kk] = vv
			t.Logf("\"%s\": \"%s\"", k+"-"+kk, vv)
		}
	}
	t.Log(m)

}
