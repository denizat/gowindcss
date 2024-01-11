package main

import (
	"io"
	"slices"
	"strings"
)

type parsedValue struct {
	name          string
	arbitraryText string
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
	var name strings.Builder
	r := strings.NewReader(s)
	var res fullClassInformation
	for {
		b, err := r.ReadByte()
		// stream ended
		if err != nil {
			if name.Len() == 0 {
				return nil
			}
			res.class = parsedValue{name: name.String(), arbitraryText: ""}
			return &res
		}
		if b == '[' {
			arb := parseArbitrary(r)
			if arb == nil {
				return nil
			}
			b, err = r.ReadByte()
			// end of string, return class
			n := name.String()
			name.Reset()
			if len(n) > 0 {
				n = n[:len(n)-1]
			}
			if err != nil {
				res.class = parsedValue{name: n, arbitraryText: *arb}
				return &res
			}
			if b == ':' {
				res.variants = append(res.variants, parsedValue{name: n, arbitraryText: *arb})
			} else {
				// malformed
				return nil
			}
		} else if b == ':' {
			res.variants = append(res.variants, parsedValue{
				name:          name.String(),
				arbitraryText: "",
			})
			name.Reset()
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
			res := v.convert(variant.arbitraryText, csses[j].CSS)
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
