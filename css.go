package main

import (
	"io"
	"strings"
)

type CSS struct {
	Selector           string
	GroupSelector      string // this is getting out of control
	PeerSelector       string // this is getting out of control
	PseudoClasses      []string
	PseudoElements     []string
	ChildCombinator    string
	MediaQueries       []string
	SupportsStatements []string
	AttributeSelectors []string
	Declarations       []CSSDeclaration
}

func CSSDeepCopy(c CSS) CSS {

	pc := make([]string, len(c.PseudoClasses))
	copy(pc, c.PseudoClasses)
	pe := make([]string, len(c.PseudoElements))
	copy(pe, c.PseudoElements)
	mq := make([]string, len(c.MediaQueries))
	copy(mq, c.MediaQueries)
	ss := make([]string, len(c.SupportsStatements))
	copy(ss, c.SupportsStatements)
	as := make([]string, len(c.AttributeSelectors))
	copy(as, c.AttributeSelectors)
	decls := make([]CSSDeclaration, len(c.Declarations))
	copy(decls, c.Declarations)

	return CSS{
		Selector:           c.Selector,
		GroupSelector:      c.GroupSelector,
		PeerSelector:       c.PeerSelector,
		PseudoClasses:      pc,
		PseudoElements:     pe,
		ChildCombinator:    c.ChildCombinator,
		MediaQueries:       mq,
		SupportsStatements: ss,
		AttributeSelectors: as,
		Declarations:       decls,
	}
}

type CSSDeclaration struct {
	Property string
	Value    string
}

type OrderedCSS struct {
	css   CSS
	order int
}

func Comparator(a OrderedCSS, b OrderedCSS) int {
	// first order by variant lexicographically
	// if equal then order by order field
	// if equal then order by baseClass lexicographically (would probably be
	// simpler to just order by the whole selector)
	return 0
}

func (c CSS) String() string {
	var b strings.Builder
	var pc string
	if len(c.PseudoClasses) >= 1 {
		pc = ":" + strings.Join(c.PseudoClasses, ":")
	}
	var pe string
	if len(c.PseudoElements) >= 1 {
		pe = "::" + strings.Join(c.PseudoElements, "::")
	}
	selector := "." + strings.Replace(c.Selector, ":", "\\:", -1)
	cc := ""
	if c.ChildCombinator != "" {
		cc = " " + c.ChildCombinator
	}
	b.WriteString(selector + cc + pc + pe + " {\n")
	for _, declaration := range c.Declarations {

		b.WriteString("\t" + declaration.Property + ": " + declaration.Value + ";\n")
	}
	b.WriteString("}\n")
	return wrapInMedias(c.MediaQueries, b.String())
}

func wrapInMedias(ms []string, s string) string {
	out := ""
	for _, m := range ms {
		out += "@media (" + m + ") {\n"
	}
	out += s
	for range ms {
		out += "}\n"
	}
	return out
}

func CSSArrToString(c []CSS) string {
	var b strings.Builder
	for _, css := range c {
		b.WriteString(css.String())
	}
	return b.String()
}
func OrderedCSSArrToString(c []OrderedCSS) string {
	var b strings.Builder
	for _, css := range c {
		b.WriteString(css.css.String())
	}
	return b.String()
}

func ParseString(s string) []OrderedCSS {
	r := strings.NewReader(s)
	csses := []OrderedCSS{}
	for {
		res := ProduceNextCSS(r)
		if res == nil {
			return csses
		}
		csses = append(csses, res...)
	}
}

func ProduceNextCSS(r io.ByteReader) []OrderedCSS {
	for {
		rf := ReadNextClass(r)
		if rf == "" {
			return nil
		}
		res := createCSSFromClassString(rf)
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

func isLowerAlnum(b byte) bool {
	return b >= 'a' || b <= 'z' || b >= '0' || b <= '9'
}

// TODO: add support for parsing arbitrary variants and values
func createCSSFromClassString(s string) []OrderedCSS {
	groups := strings.Split(s, ":")
	css, ok := baseClasses[groups[len(groups)-1]]
	css.css.Selector = s
	if !ok {
		return nil
	}
	csses := []OrderedCSS{css}
	for i := len(groups) - 2; i >= 0; i-- {
		l := len(csses)
		for j := 0; j < l; j++ {
			group := groups[i]
			css := csses[j]
			variant := variants[group]
			if variant == nil {
				return nil
			}
			gen := variant.convert(nil, css.css)
			if len(gen) == 0 {
				return nil
			}
			if len(gen) >= 1 {
				csses[j].css = gen[0]
			}
			if len(gen) == 2 {
				csses = append(csses, OrderedCSS{gen[1], csses[j].order})
			}
		}
	}
	return csses
}
