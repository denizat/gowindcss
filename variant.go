package main

// https://tailwindcss.com/docs/hover-focus-and-other-states
type Variant interface {
	convert(arbitaryValue *string, c CSS) []CSS
	base() string
}

func MakeVariants(c *Config) map[string]Variant {
	return concatMaps(
		variantMapFromArrs(pseudoClassVariants),
		variantMapFromArrs(pseudoElementVariants),
		variantMapFromArrs([]DoublePsuedoElementVariant{markerVariant}),
		variantMapFromArrs(genBreakpointsVariant(c)),
	)
}

var variants = MakeVariants(nil)

func variantMapFromArrs[T Variant](arr []T) map[string]Variant {
	m := map[string]Variant{}
	for _, v := range arr {
		m[v.base()] = v
	}
	return m
}

type PseudoClassVariant struct {
	name string
}

func (p PseudoClassVariant) convert(arbitaryValue *string, c CSS) []CSS {
	if arbitaryValue != nil {
		panic("expected arbitrary value to be nil")
	}
	c.PseudoClasses = append(c.PseudoClasses, p.name)
	return []CSS{c}
}
func (p PseudoClassVariant) base() string {
	return p.name
}

var pseudoClassVariants = []PseudoClassVariant{
	{"hover"},
	{"focus"},
}

type PseudoElementVariant struct {
	name  string
	value string
}

func PseudoElementNameSameAsValue(n string) PseudoElementVariant {
	return PseudoElementVariant{name: n, value: n}
}

func (p PseudoElementVariant) convert(arbitaryValue *string, c CSS) []CSS {
	if arbitaryValue != nil {
		panic("expected arbitrary value to be nil")
	}
	c.PseudoElements = append(c.PseudoClasses, p.name)
	return []CSS{c}
}
func (p PseudoElementVariant) base() string {
	return p.name
}

var pseudoElementVariants = []PseudoElementVariant{

	PseudoElementNameSameAsValue("before"),
	PseudoElementNameSameAsValue("after"),
	PseudoElementNameSameAsValue("placeholder"),
	{"file", "file-selector-button"},
	PseudoElementNameSameAsValue("first-line"),
	PseudoElementNameSameAsValue("backdrop"),
}

type DoublePsuedoElementVariant struct{ name string }

func (m DoublePsuedoElementVariant) convert(arbitraryValue *string, c CSS) []CSS {
	c.PseudoElements = append(c.PseudoElements, m.name)
	cc := CSSDeepCopy(c)
	cc.ChildCombinator = "*"
	return []CSS{c, cc}
}
func (m DoublePsuedoElementVariant) base() string {
	return m.name
}

var markerVariant = DoublePsuedoElementVariant{"marker"}

type MediaVariant struct {
	name  string
	value string
}

func (m MediaVariant) convert(arbitraryValue *string, c CSS) []CSS {
	c.MediaQueries = append(c.PseudoElements, m.value)
	return []CSS{c}
}
func (m MediaVariant) base() string {
	return m.name
}

// preferences and other things
var preferences = []MediaVariant{
	{"dark", "prefers-color-scheme: dark"},
	{"motion-reduce", "prefers-reduced-motion: reduce"},
	{"contrast-more", "prefers-contrast: more"},

	{"forced-colors", "forced-colors: active"},
	{"portrait", "orientation: portrait"},
	{"landscape", "orientation: landscape"},
	// MEDIA PRINT IS DIFFERENT FROM THE REST
	// @media print
	// instead of @media(print)
	{"print", "print"},
}

type supportsVariant struct{}

// does not support checking properties yet
func (s supportsVariant) convert(arbitraryValue *string, c CSS) []CSS {
	c.SupportsStatements = append(c.SupportsStatements, *arbitraryValue)
	return []CSS{c}
}
func (s supportsVariant) base() string {
	return "supports"
}

type customSupportsVariant struct {
	name  string
	value string
}

func (s customSupportsVariant) convert(arbitraryValue *string, c CSS) []CSS {
	c.SupportsStatements = append(c.SupportsStatements, s.value)
	return []CSS{c}
}
func (s customSupportsVariant) base() string {
	return "supports-" + s.name
}

type ariaStatesVariant struct {
	name  string
	value string
}

type BreakpointsVariant struct {
	name  string
	value string
}

func genBreakpointsVariant(_ *Config) []Variant {
	arr := []Variant{}
	for k, v := range defaultBreakpoints {
		arr = append(arr, BreakpointsVariant{name: k, value: v})
	}
	return arr
}

func (b BreakpointsVariant) convert(arbitraryValue *string, c CSS) []CSS {
	if arbitraryValue != nil {
		return nil
	}
	c.MediaQueries = append(c.MediaQueries, "min-width: "+b.value)
	return []CSS{c}
}
func (b BreakpointsVariant) base() string {
	return b.name
}
