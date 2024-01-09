package main

import "testing"

func TestThing(t *testing.T) {
	csses := ParseString("marker:aspect-video")
	s := OrderedCSSArrToString(csses)
	t.Logf(s)
}
