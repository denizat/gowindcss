package main

import (
	"io"
	"strings"
)

type parsedVariant struct {
	name          string
	arbitraryText string // use the zero value lol
}
type parsedClass struct {
	name          string
	arbitraryText string
}
type fullClassInformation struct {
	variants []parsedVariant
	class    parsedClass
}

func parse(s string) *fullClassInformation {
	return nil
}

func generate(f fullClassInformation, vs VariantMap, bs BaseClassMap) []OrderedCSS {
	return nil
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
		rf := ReadNextClass(r)
		if rf == "" {
			return nil
		}
		res := createCSSFromClassString(rf, vs, bs)
		if res != nil {
			return res
		}
	}
}

// TODO: add support for parsing variants
func ReadNextClass(r io.ByteReader) string {
	var bd strings.Builder
	writeStarted := false
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			break
		}
		if isLowerAlnum(b) || b == '-' || b == ':' {
			bd.WriteByte(b)
			writeStarted = true
		} else if writeStarted {
			break
		}
	}
	return bd.String()
}

// TODO: add support for parsing arbitrary variants and values
func createCSSFromClassString(s string, vs map[string]Variant, bs map[string]OrderedCSS) []OrderedCSS {
	groups := strings.Split(s, ":")
	base := groups[len(groups)-1]
	css, ok := bs[base]
	if !ok {
		parts := strings.Split(base, "-[")
		if len(parts) < 2 || parts[1] == "" {
			return nil
		}
		key := strings.Join(parts[:len(parts)-1], "") // this join should be unnecessary
		arbitraryValue := parts[len(parts)-1]
		if len(arbitraryValue) == 0 {
			return nil
		}
		arbitraryValue = arbitraryValue[:len(arbitraryValue)-1]
		baseClass, ok := baseClassesArbitrary[key]
		if !ok {
			return nil
		}
		css = baseClass.arbitraryValue(arbitraryValue)
	}

	css.Selector = s
	csses := []OrderedCSS{css}
	for i := len(groups) - 2; i >= 0; i-- {
		l := len(csses)
		for j := 0; j < l; j++ {
			group := groups[i]
			css := csses[j]
			variant := vs[group]
			if variant == nil {
				return nil
			}
			gen := variant.convert(nil, css.CSS)
			if len(gen) == 0 {
				return nil
			}
			if len(gen) >= 1 {
				csses[j].CSS = gen[0]
			}
			if len(gen) == 2 {
				csses = append(csses, OrderedCSS{gen[1], csses[j].order})
			}
		}
	}
	return csses
}
