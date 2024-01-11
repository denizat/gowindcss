package main

import (
	"io"
	"slices"
	"strings"
)

type parsedValue struct {
	name          string
	arbitraryText string
	slashText     string
}
type fullClassInformation struct {
	variants []parsedValue
	class    parsedValue
}

func parse(s string) *fullClassInformation {
	return parsestr(s)
}

func ParseString(s string, vs map[string]Variant, bs map[string]OrderedCSS) []OrderedCSS {
	r := strings.NewReader(s)
	csses := []OrderedCSS{}
	for {
		res := ProduceNextCSS(r, vs, bs)
		if res == nil {
			return csses
		}
		csses = append(csses, res...)
	}
}

func ProduceNextCSS(r io.ByteReader, vs map[string]Variant, bs map[string]OrderedCSS) []OrderedCSS {
	for {
		s := grabFirstPossibleValidString(r)
		if len(s) == 0 {
			return nil
		}
		res := parsestr(s)
		if res == nil {
			continue
		}
		csses := createCSSFromClassInformation(*res, s, vs, bs)
		if csses != nil {
			return csses
		}
	}

}

func grabFirstPossibleValidString(r io.ByteReader) string {
	var sb strings.Builder
	for {
		b, err := r.ReadByte()
		if err != nil || b == ' ' || b == '\n' || b == '\t' || b == '\'' || b == '"' || b == '`' {
			return sb.String()
		}
		sb.WriteByte(b)
	}
}

func parsestr(s string) *fullClassInformation {
	r := strings.NewReader(s)
	var res fullClassInformation
	for {
		pv, variant := parseNextPart(r)
		if pv == nil {
			return nil
		}
		if variant {
			res.variants = append(res.variants, *pv)
		} else {
			res.class = *pv
			return &res
		}
	}
}

// returns the parsed value and if it was a variant or not
// returns nil if unable to parse, if it returns nil then the bool does not matter
func parseNextPart(r io.ByteReader) (*parsedValue, bool) {
	// I might be able to flatten this function by storing some state
	// and updating all the top level if statements to check that state first
	var name strings.Builder
	for {
		b, err := r.ReadByte()
		if err != nil {
			return &parsedValue{name: name.String()}, false
		}
		if b == ':' {
			return &parsedValue{name: name.String()}, true
		}
		// I do not know how to flatten this
		if b == '[' {
			res := parseArbitrary(r)
			if res == nil {
				return nil, false
			}
			// next read must be EOF or slash or colon
			b, err := r.ReadByte()
			n := name.String()
			if len(n) > 0 {
				n = n[:len(n)-1]
			}
			pv := &parsedValue{name: n, arbitraryText: *res}
			if err != nil {
				return pv, false
			}
			if b == ':' {
				return pv, true
			}
			if b == '/' {
				slash, colon := parseSlash(r)
				if slash == "" {
					return nil, false
				}
				pv.slashText = slash
				if colon {
					return pv, true
				} else {
					return pv, false
				}
			}
		} else if b == '/' {
			slash, colon := parseSlash(r)
			if slash == "" {
				return nil, false
			}
			pv := &parsedValue{name: name.String(), slashText: slash}
			if colon {
				return pv, true
			} else {
				return pv, false
			}
		} else {
			name.WriteByte(b)
		}
	}
}

// [abc] returns "abc"
// [] returns ""
// [asdfasdfhalksjdfhaslkjdfhasdkjf (never closed) returns nil
// only returns nil if there is a problem with the arbitrary thing
// TODO: handle edge case of nested []s, probably use some sort of escaping, figure out how tailwind guys do it
func parseArbitrary(r io.ByteReader) *string {
	var sb strings.Builder
	for {
		b, err := r.ReadByte()
		// TODO: handle this error better, propagate it in the future
		if err != nil {
			return nil
		}
		if b == ']' {
			s := sb.String()
			return &s
		}
		sb.WriteByte(b)
	}
}

// only stops at colons
// returns the text after the slash and if it ended with a colon
// if it returns false that means the stream ended
func parseSlash(r io.ByteReader) (string, bool) {
	var sb strings.Builder
	for {
		b, err := r.ReadByte()
		// end of string just return
		if err != nil {
			return sb.String(), false
		}
		if b == ':' {
			return sb.String(), true
		}
		sb.WriteByte(b)
	}
}

func createCSSFromClassInformation(c fullClassInformation, selector string, vs map[string]Variant, bs map[string]OrderedCSS) []OrderedCSS {
	var css OrderedCSS
	if c.class.arbitraryText == "" {
		val, ok := bs[c.class.name]
		if !ok {
			return nil
		}
		css = val
	} else {
		arb, ok := baseClassesArbitrary[c.class.name]
		if !ok {
			return nil
		}
		css = arb.arbitraryValue(c.class.arbitraryText)
	}
	css.Selector = selector
	csses := []OrderedCSS{css}
	slices.Reverse(c.variants)
	for _, variant := range c.variants {
		v, ok := vs[variant.name]
		if !ok {
			return nil
		}
		l := len(csses)
		for j := 0; j < l; j++ {
			res := v.convert(variant.arbitraryText, "", csses[j].CSS)
			if res == nil {
				return nil
			}
			if len(res) == 0 {
				return nil
			}
			if len(res) >= 1 {
				csses[j].CSS = res[0]
			}
			if len(res) == 2 {
				csses = append(csses, OrderedCSS{res[1], csses[j].order})
			}
		}
	}
	return csses
}
