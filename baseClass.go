package main

type BaseClass interface {
	arbitraryValue(v string) *OrderedCSS
	produceMap( /* TODO: add optional customizations */ ) map[string]OrderedCSS
	baseForArbitraryValue() *string
}

var baseClasses = concatMaps(aspectRatio.produceMap(), columns.produceMap())

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

func (a ArbitraryValueKeywordClass) arbitraryValue(v string) *OrderedCSS {
	return &OrderedCSS{
		css: CSS{
			Selector:     a.name + "-[" + v + "]",
			Declarations: []CSSDeclaration{{a.property, v}},
		},
		order: a.order,
	}
}
func (a ArbitraryValueKeywordClass) produceMap() map[string]OrderedCSS {
	m := map[string]OrderedCSS{}
	for k, v := range a.defaults {
		n := a.name + "-" + k
		m[n] = OrderedCSS{
			css: CSS{
				Selector:     n,
				Declarations: []CSSDeclaration{{a.property, v}},
			},
			order: 0,
		}
	}
	return m
}

func (a ArbitraryValueKeywordClass) baseForArbitraryValue() *string {
	return &a.name
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
