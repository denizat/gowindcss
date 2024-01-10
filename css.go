package main

import (
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
type CSSDeclaration struct {
	Property string
	Value    string
}

type OrderedCSS struct {
	CSS
	order int
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

func OrderedCSSDeepCopy(c OrderedCSS) OrderedCSS {
	copy := CSSDeepCopy(c.CSS)
	return OrderedCSS{copy, c.order}
}

func OrderedCSSLess(a, b OrderedCSS) int {
	if a.order < b.order {
		return -1
	}
	if a.order > b.order {
		return 1
	}
	as := a.Selector
	bs := b.Selector
	if len(as) < len(bs) {
		return -1
	}
	if len(as) > len(bs) {
		return 1
	}
	if as < bs {
		return -1
	}
	if as > bs {
		return 1
	}
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
	selector := strings.Replace(c.Selector, "[", "\\[", -1)
	selector = strings.Replace(selector, "]", "\\]", -1)
	selector = "." + strings.Replace(selector, ":", "\\:", -1)
	cc := ""
	if c.ChildCombinator != "" {
		cc = " " + c.ChildCombinator
	}
	b.WriteString(selector + cc + pc + pe + " {\n")
	for _, declaration := range c.Declarations {

		b.WriteString(indent + declaration.Property + ": " + declaration.Value + ";\n")
	}
	b.WriteString("}\n")
	return wrapInMedias(c.MediaQueries, b.String())
}

const indent = "  "

func wrapInMedias(ms []string, s string) string {
	out := ""
	for i, m := range ms {
		out += strings.Repeat(indent, i) + "@media (" + m + ") {\n"
	}
	lines := strings.Split(s, "\n")
	for i := range lines[:len(lines)-1] {
		out += strings.Repeat(indent, len(ms)) + lines[i]
		if i < len(lines)-1 {
			out += "\n"
		}
	}
	for i := range ms {
		out += strings.Repeat(indent, len(ms)-i-1) + "}\n"
	}
	return out
}
func OrderedCSSArrToString(c []OrderedCSS) string {
	var b strings.Builder
	for _, css := range c {
		b.WriteString(css.String())
	}
	return b.String()
}
