package main

import (
	"fmt"
	"strconv"
)

type BaseClass interface {
	produceMap(config map[string]any) map[string]OrderedCSS
}
type ArbitraryValue interface {
	arbitraryValue(v string) OrderedCSS
	baseForArbitraryValue() string
}

type Options struct {
	Colors     map[string]string `json:"colors"`
	FontFamily []string          `json:"fontFamily"`
	Screens    map[string]string `json:"screens"`
	Spacing    map[string]string `json:"spacing"`
}

type Theme struct {
	Options
	Extend Options `json:"extend"`
}

type Config struct {
	Theme Theme `json:"theme"`
}

func MakeBaseClasses(config *Config) map[string]OrderedCSS {
	return concatMaps(
		// Layout
		aspectRatio.produceMap(config),
		// container is unique
		columns.produceMap(config),
		breakAfter.produceMap(config),
		breakBefore.produceMap(config),
		breakInside.produceMap(config),
		boxDecoration.produceMap(config),
		boxSizing.produceMap(config),
		display.produceMap(config),
		floats.produceMap(config),
		clear.produceMap(config),
		isolation.produceMap(config),
		// Flexbox & Grid
		flexBasis.produceMap(config),
		grow.produceMap(config),
		// Backgrounds
		backgroundColor.produceMap(config),
		// Typography
		textColor.produceMap(config),
		textWrap.produceMap(config),
	)
}

var baseClassesArbitrary = map[string]ArbitraryValue{
	aspectRatio.baseForArbitraryValue():     aspectRatio,
	columns.baseForArbitraryValue():         columns,
	flexBasis.baseForArbitraryValue():       flexBasis,
	grow.baseForArbitraryValue():            grow,
	backgroundColor.baseForArbitraryValue(): backgroundColor,
	textColor.baseForArbitraryValue():       textColor,
}

// TODO: reorder this to fit what tailwind does
const (
	aspectRatioOrder = 100 * iota
	columnsOrder
)
const _ = (iota+4)*columnsOrder + 4

type ArbitraryValueKeywordClass struct {
	name     string
	property string
	defaults map[string]string
	order    int
}

func (a ArbitraryValueKeywordClass) arbitraryValue(v string) OrderedCSS {
	return OrderedCSS{
		css: CSS{
			Declarations: []CSSDeclaration{{a.property, v}},
		},
		order: a.order,
	}
}
func (a ArbitraryValueKeywordClass) produceMap(config *Config) map[string]OrderedCSS {
	m := map[string]OrderedCSS{}
	for k, v := range a.defaults {
		n := a.name
		if k != "" {
			n += "-" + k
		}
		m[n] = OrderedCSS{
			css: CSS{
				Declarations: []CSSDeclaration{{a.property, v}},
			},
			order: a.order,
		}
	}
	return m
}

func (a ArbitraryValueKeywordClass) baseForArbitraryValue() string {
	return a.name
}

var grow = ArbitraryValueKeywordClass{
	name:     "grow",
	property: "flex-grow",
	defaults: map[string]string{
		"":  "1",
		"0": "0",
	},
	order: 0,
}

var aspectRatio = ArbitraryValueKeywordClass{
	name:     "aspect",
	property: "aspect-ratio",
	defaults: map[string]string{
		"auto":   "auto",
		"square": "1 / 1",
		"video":  "16 / 9",
	},
	order: aspectRatioOrder,
}

var columns = ArbitraryValueKeywordClass{
	name:     "columns",
	property: "columns",
	defaults: map[string]string{
		"1":    "1",
		"2":    "2",
		"3":    "3",
		"4":    "4",
		"5":    "5",
		"6":    "6",
		"7":    "7",
		"8":    "8",
		"9":    "9",
		"10":   "10",
		"11":   "11",
		"12":   "12",
		"auto": "auto",
		"3xs":  "16rem",
		"2xs":  "18rem",
		"xs":   "20rem",
		"sm":   "24rem",
		"md":   "28rem",
		"lg":   "32rem",
		"xl":   "36rem", // tailwind does more, do later
	},
	order: columnsOrder,
}

type KeywordBaseClass struct {
	name     string
	property string
	values   map[string]string
	order    int
}

// copied from the arbitrary value one
func (a KeywordBaseClass) produceMap(config *Config) map[string]OrderedCSS {
	m := map[string]OrderedCSS{}
	for k, v := range a.values {
		n := a.name + "-" + k
		m[n] = OrderedCSS{
			css: CSS{
				Declarations: []CSSDeclaration{{a.property, v}},
			},
			order: a.order,
		}
	}
	return m
}

var breakAfter = KeywordBaseClass{
	name:     "break-after",
	property: "break-after",
	values: map[string]string{
		"auto":       "auto",
		"avoid":      "avoid",
		"all":        "all",
		"avoid-page": "avoid-page",
		"page":       "page",
		"left":       "left",
		"right":      "right",
		"column":     "column",
	},
	order: 0,
}
var breakBefore = KeywordBaseClass{
	name:     "break-before",
	property: "break-before",
	values: map[string]string{
		"auto":       "auto",
		"avoid":      "avoid",
		"all":        "all",
		"avoid-page": "avoid-page",
		"page":       "page",
		"left":       "left",
		"right":      "right",
		"column":     "column",
	},
	order: 0,
}
var breakInside = KeywordBaseClass{
	name:     "break-inside",
	property: "break-inside",
	values: map[string]string{
		"auto":         "auto",
		"avoid":        "avoid",
		"avoid-page":   "avoid-page",
		"avoid-column": "avoid-column",
	},
	order: 0,
}
var boxDecoration = KeywordBaseClass{
	name:     "box-decoration",
	property: "box-decoration-break",
	values: map[string]string{
		"clone": "clone",
		"slice": "slice",
	},
	order: 0,
}
var boxSizing = KeywordBaseClass{
	name:     "box",
	property: "box-sizing",
	values: map[string]string{
		"border":  "border-box",
		"content": "content-box",
	},
	order: 0,
}

var floats = KeywordBaseClass{
	name:     "float",
	property: "float",
	values: map[string]string{
		"start": "inline-start",
		"end":   "inline-end",
		"left":  "left",
		"right": "right",
		"none":  "none",
	},
	order: 0,
}
var clear = KeywordBaseClass{
	name:     "float",
	property: "float",
	values: map[string]string{
		"start": "inline-start",
		"end":   "inline-end",
		"left":  "left",
		"right": "right",
		"both":  "both",
		"none":  "none",
	},
	order: 0,
}

var textWrap = KeywordBaseClass{
	// TODO: check for duplicates
	name:     "text",
	property: "text-wrap",
	values: map[string]string{
		"wrap":    "wrap",
		"nowrap":  "nowrap",
		"balance": "balance",
		"pretty":  "pretty",
	},
	order: 0,
}

// TODO: replace all instances of order: 0

type RandomKeywordBaseClass struct {
	keywords map[string]string
	property string
	order    int
}

func (r RandomKeywordBaseClass) produceMap(config *Config) map[string]OrderedCSS {
	m := map[string]OrderedCSS{}
	for k, v := range r.keywords {
		m[k] = OrderedCSS{
			css: CSS{
				Declarations: []CSSDeclaration{{Property: r.property, Value: v}},
			},
			order: r.order,
		}
	}
	return m
}

var display = RandomKeywordBaseClass{
	keywords: map[string]string{
		"block":              "block",
		"inline-block":       "inline-block",
		"inline":             "inline",
		"flex":               "flex",
		"inline-flex":        "inline-flex",
		"table":              "table",
		"inline-table":       "inline-table",
		"table-caption":      "table-caption",
		"table-cell":         "table-cell",
		"table-column":       "table-column",
		"table-column-group": "table-column-group",
		"table-footer-group": "table-footer-group",
		"table-header-group": "table-header-group",
		"table-row-group":    "table-row-group",
		"table-row":          "table-row",
		"flow-root":          "flow-root",
		"grid":               "grid",
		"inline-grid":        "inline-grid",
		"contents":           "contents",
		"list-item":          "list-item",
		// the exception behind the duplication
		"hidden": "none",
	},
	property: "display",
	order:    0,
}

var isolation = RandomKeywordBaseClass{
	keywords: map[string]string{
		"isolate":        "isolate",
		"isolation-auto": "auto",
	},
	property: "isolation",
	order:    0,
}

type ArbitraryNumericalBaseClass struct {
	name     string
	property string
	keywords map[string]string
	order    int
}

func (a ArbitraryNumericalBaseClass) produceMap(config *Config) map[string]OrderedCSS {
	m := map[string]OrderedCSS{}
	for _, num := range defaultNums {
		n, err := strconv.ParseFloat(num, 64)
		n = n / 4
		// should not happen
		if err != nil {
			continue
		}
		m[a.name+"-"+num] = OrderedCSS{
			css: CSS{
				Declarations: []CSSDeclaration{{Property: a.property, Value: fmt.Sprint(n) + "rem"}},
			},
			order: a.order,
		}
	}
	for _, fraction := range defaultFractions {
		n, err := parseFraction(fraction)
		// should not happen
		if err != nil {
			continue
		}
		m[a.name+"-"+fraction] = OrderedCSS{
			css: CSS{
				Declarations: []CSSDeclaration{{Property: a.property, Value: fmt.Sprint(n) + "%"}},
			},
			order: a.order,
		}
	}
	for k, v := range a.keywords {
		m[a.name+"-"+k] = OrderedCSS{
			css:   CSS{Declarations: []CSSDeclaration{{Property: a.property, Value: v}}},
			order: a.order,
		}
	}
	return m
}

func (a ArbitraryNumericalBaseClass) arbitraryValue(v string) OrderedCSS {
	return OrderedCSS{
		css:   CSS{Declarations: []CSSDeclaration{{Property: a.property, Value: v}}},
		order: a.order,
	}
}
func (a ArbitraryNumericalBaseClass) baseForArbitraryValue() string {
	return a.name
}

var flexBasis = ArbitraryNumericalBaseClass{
	name:     "basis",
	property: "flex-basis",
	keywords: map[string]string{
		"auto": "auto",
		"full": "full",
	},
	order: 0,
}

type ArbitraryColorBaseClass struct {
	name     string
	property string
	order    int
}

func (a ArbitraryColorBaseClass) produceMap(config *Config) map[string]OrderedCSS {
	m := map[string]OrderedCSS{}
	for k, v := range defaultColors {
		m[a.name+"-"+k] = OrderedCSS{
			css: CSS{
				Declarations: []CSSDeclaration{{Property: a.property, Value: v}},
			},
			order: a.order,
		}
	}
	return m
}
func (a ArbitraryColorBaseClass) baseForArbitraryValue() string {
	return a.name
}
func (a ArbitraryColorBaseClass) arbitraryValue(v string) OrderedCSS {
	return OrderedCSS{
		css:   CSS{Declarations: []CSSDeclaration{{Property: a.property, Value: v}}},
		order: a.order,
	}
}

var backgroundColor = ArbitraryColorBaseClass{
	name:     "bg",
	property: "background-color",
	order:    0,
}
var textColor = ArbitraryColorBaseClass{
	name:     "text",
	property: "color",
	order:    0,
}
