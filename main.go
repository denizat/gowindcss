package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
)

//go:embed main.go main_test.go go.mod go.sum tests/* README.md LICENSE
var sourceCode embed.FS

// TODO: remove as many panics as I can and do real error handling
func main() {
	var configFileName string
	flag.StringVar(&configFileName, "c", "", "location of a configuration file, optional")
	var dump bool
	flag.BoolVar(&dump, "d", false, "dump all css generated from base classes")
	formatFile := flag.String("f", "", "format the file")
	// repl that can print out what class names will become and prints out their order number, but also
	// can do formatting if you start the line with "
	repl := flag.Bool("r", false, "interactive mode")
	_ = repl
	list := flag.Bool("l", false, "list out base classes in order and the declarations they will generate")
	_ = flag.String("i", "", "input css file")
	_ = flag.String("o", "", "output css file")                                                                           // or could do redirection?
	writeSource := flag.Bool("writeSource", false, "write the entire source code of this program to ./gowindcss-source/") // or could do redirection?
	flag.Parse()

	if *writeSource {
		fileNames := []string{
			"main.go", "main_test.go", "go.mod", "go.sum",
			"tests/config.json", "tests/configtests.txt", "tests/defaultstests.txt", "tests/formattertests.txt",
			"README.md", "LICENSE"}
		// figure out what these perms mean I forgot lol
		folder := "gowindcss-source"
		err := os.Mkdir(folder, 0777)
		if err != nil {
			panic(err)
		}
		err = os.Mkdir(folder+"/tests", 0777)
		if err != nil {
			panic(err)
		}

		for _, fileName := range fileNames {
			f, err := sourceCode.ReadFile(fileName)
			if err != nil {
				// SHOULD NEVER EVER BE HERE
				fmt.Printf("sorry, someone made a silly mistake and for some reason we are unable to open the file: `%s`\n", fileName)
			} else {
				out, err := os.Create(folder + "/" + fileName)
				if err != nil {
					panic(err)
				}
				out.Write(f)
			}
		}
		os.Exit(0)
	}

	var bs map[string]OrderedCSS
	if configFileName != "" {
		bs = HandleConfigFile(&configFileName)
	} else {
		bs = MakeBaseClasses(nil)
	}

	if dump {
		ks := []OrderedCSS{}
		for k, v := range bs {
			v.Selector = k
			ks = append(ks, v)
		}
		slices.SortFunc(ks, func(a, b OrderedCSS) int {
			return OrderedCSSLess(a, b)
		})
		fmt.Println(OrderedCSSArrToString(ks))
		os.Exit(0)
	}

	if *list {
		ks := []OrderedCSS{}
		for k, v := range bs {
			v.Selector = k
			ks = append(ks, v)
		}
		slices.SortFunc(ks, func(a, b OrderedCSS) int {
			return OrderedCSSLess(a, b)
		})
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "class\torder number\tdeclarations\t")
		for _, c := range ks {
			fmt.Fprintf(w, "%s\t%d\t", c.Selector, c.order)
			for _, d := range c.Declarations {
				fmt.Fprintf(w, "%s: %s;\t", d.Property, d.Value)
			}
			fmt.Fprintln(w)
		}
		w.Flush()
		os.Exit(0)
	}

	if *formatFile != "" {
		inFile, err := os.Open(*formatFile)
		if err != nil {
			panic(err)
		}
		outFile, err := os.OpenFile(*formatFile, os.O_RDWR, 0777)
		if err != nil {
			panic(err)
		}
		reader := bufio.NewReader(inFile)
		writer := bufio.NewWriter(outFile)
		Format(reader, writer, variants, bs)
		os.Exit(0)
	}

	// do the rest
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	for scanner.Scan() {
		csses := ParseString(scanner.Text(), variants, bs)
		s := OrderedCSSArrToString(csses)
		_, err := writer.WriteString(s)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
		}
		err = writer.Flush()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
		}
	}
}

func ParseConfigFile(file []byte) map[string]OrderedCSS {
	var config Config
	json.Unmarshal(file, &config)
	if config.Theme.Colors != nil {
		defaultColors = config.Theme.Colors
	}
	if config.Theme.Extend.Colors != nil {
		maps.Copy(defaultColors, config.Theme.Extend.Colors)
	}
	if config.Theme.Screens != nil {
		defaultBreakpoints = config.Theme.Screens
	}
	if config.Theme.Extend.Screens != nil {
		maps.Copy(defaultColors, config.Theme.Extend.Screens)
	}
	return MakeBaseClasses(&config)

}

func HandleConfigFile(fileName *string) map[string]OrderedCSS {
	if fileName == nil {
		return MakeBaseClasses(nil)
	}
	bs, err := os.ReadFile(*fileName)
	if err != nil {
		panic(err)
	}
	return ParseConfigFile(bs)
}

func Format(r io.ByteReader, w io.ByteWriter, vs map[string]Variant, bs map[string]OrderedCSS) {
	for {
		streamUntilMatch(r, w, "class=\"")
		s, err := collectUntil(r, '"')
		if err != nil {
			break
		}
		s = realFormatGiveBetterName(s, vs, bs) + "\""
		for i := range s {
			w.WriteByte(s[i])
		}
	}
}

func collectUntil(r io.ByteReader, c byte) (string, error) {
	var sb strings.Builder
	for {
		b, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		if b == c {
			return sb.String(), nil
		}
		sb.WriteByte(b)
	}
}

func streamUntilMatch(r io.ByteReader, w io.ByteWriter, s string) {
	i := 0
	for {
		b, err := r.ReadByte()
		if err != nil {
			return
		}
		// TODO: handle this error better
		if w.WriteByte(b) != nil {
			return
		}
		if b == s[i] {
			i++
			if i == len(s) {
				return
			}
		} else if i > 0 {
			i = 0
		}
	}
}

func realFormatGiveBetterName(s string, vs map[string]Variant, bs map[string]OrderedCSS) string {
	classNames := strings.Split(s, " ")
	slices.SortFunc(classNames, func(a, b string) int {
		ac := ParseString(a, vs, bs)[0]
		bc := ParseString(b, vs, bs)[0]
		return OrderedCSSLess(ac, bc)
	})
	return strings.Join(classNames, " ")
}

//////////////////////////////////////////// CSS

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

//////////////////////////////////////////// ENGINE

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
		if err != nil || b == ' ' || b == '\n' || b == '\t' || b == '"' || b == '`' {
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

//////////////////////////////////////////// BASE CLASSES

type BaseClassMap map[string]OrderedCSS

type BaseClass interface {
	produceMap(config *Config) map[string]OrderedCSS
}
type ArbitraryValueClass interface {
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
		flexDirection.produceMap(config),
		flexWrap.produceMap(config),
		grow.produceMap(config),
		// Backgrounds
		backgroundColor.produceMap(config),
		// Typography
		textColor.produceMap(config),
		textWrap.produceMap(config),
	)
}

var baseClassesArbitrary = map[string]ArbitraryValueClass{
	totallyArbitraryLOL.baseForArbitraryValue(): totallyArbitraryLOL,
	aspectRatio.baseForArbitraryValue():         aspectRatio,
	columns.baseForArbitraryValue():             columns,
	flexBasis.baseForArbitraryValue():           flexBasis,
	grow.baseForArbitraryValue():                grow,
	backgroundColor.baseForArbitraryValue():     backgroundColor,
	textColor.baseForArbitraryValue():           textColor,
}

// TODO: reorder this to fit what tailwind does
const (
	_ = iota * 100

	aspectRatioOrder
	columnsOrder
	breakBeforeOrder
	breakInsideOrder
	breakAfterOrder
	textWrapOrder

	growOrder
	flexDirectionOrder
	totallyArbitraryLOLOrder
)

const (
	keywordsOrder = iota
	numbersOrder
	fractionsOrder
)

type ArbitraryValueKeywordClass struct {
	name     string
	property string
	defaults map[string]string
	order    int
}

func (a ArbitraryValueKeywordClass) arbitraryValue(v string) OrderedCSS {
	return OrderedCSS{
		CSS{Declarations: []CSSDeclaration{{a.property, v}}},
		a.order,
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
			CSS{Declarations: []CSSDeclaration{{a.property, v}}},
			a.order,
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
	order: growOrder,
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

type LOLOLOL struct{}

func (L LOLOLOL) arbitraryValue(v string) OrderedCSS {
	declsStrs := strings.Split(v, ";")
	decls := []CSSDeclaration{}
	for _, s := range declsStrs {
		parts := strings.Split(s, ":")
		if len(parts) == 2 {

			decls = append(decls, CSSDeclaration{parts[0], parts[1]})
		} else {
			decls = append(decls, CSSDeclaration{parts[0], ""})
		}
	}
	return OrderedCSS{
		CSS{Declarations: decls}, totallyArbitraryLOLOrder,
	}
}

func (L LOLOLOL) baseForArbitraryValue() string {
	return ""
}

var totallyArbitraryLOL = LOLOLOL{}

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
			CSS{Declarations: []CSSDeclaration{{a.property, v}}},
			a.order,
		}
	}
	return m
}

var flexWrap = KeywordBaseClass{
	name:     "flex",
	property: "flex-wrap",
	values: map[string]string{
		"wrap":         "wrap",
		"wrap-reverse": "wrap-reverse",
		"nowrap":       "nowrap",
	},
	order: 0,
}

var flexDirection = KeywordBaseClass{
	name:     "flex",
	property: "flex-direction",
	values: map[string]string{
		"row":         "row",
		"row-reverse": "row-reverse",
		"col":         "col",
		"col-reverse": "col-reverse",
	},
	order: flexDirectionOrder,
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
	order: breakAfterOrder,
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
	order: breakBeforeOrder,
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
	order: breakInsideOrder,
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
	order: textWrapOrder,
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
			CSS{Declarations: []CSSDeclaration{{Property: r.property, Value: v}}},
			r.order,
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
			CSS{Declarations: []CSSDeclaration{{Property: a.property, Value: fmt.Sprint(n) + "rem"}}},
			a.order + numbersOrder,
		}
	}
	for _, fraction := range defaultFractions {
		n, err := parseFraction(fraction)
		// should not happen
		if err != nil {
			continue
		}
		m[a.name+"-"+fraction] = OrderedCSS{
			CSS{Declarations: []CSSDeclaration{{Property: a.property, Value: fmt.Sprint(n) + "%"}}},
			a.order + fractionsOrder,
		}
	}
	for k, v := range a.keywords {
		m[a.name+"-"+k] = OrderedCSS{
			CSS{Declarations: []CSSDeclaration{{Property: a.property, Value: v}}},
			a.order + keywordsOrder,
		}
	}
	return m
}

func (a ArbitraryNumericalBaseClass) arbitraryValue(v string) OrderedCSS {
	return OrderedCSS{
		CSS{Declarations: []CSSDeclaration{{Property: a.property, Value: v}}},
		a.order,
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
			CSS{Declarations: []CSSDeclaration{{Property: a.property, Value: v}}},
			a.order,
		}
	}
	return m
}
func (a ArbitraryColorBaseClass) baseForArbitraryValue() string {
	return a.name
}
func (a ArbitraryColorBaseClass) arbitraryValue(v string) OrderedCSS {
	return OrderedCSS{
		CSS{Declarations: []CSSDeclaration{{Property: a.property, Value: v}}},
		a.order,
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

//////////////////////////////////////////// VARIANTS

type VariantMap map[string]Variant

// https://tailwindcss.com/docs/hover-focus-and-other-states
type Variant interface {
	convert(arbitaryValue, slashText string, c CSS) []CSS
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

func (p PseudoClassVariant) convert(arbitaryValue string, slashText string, c CSS) []CSS {
	c.PseudoClasses = append(c.PseudoClasses, p.name)
	return []CSS{c}
}
func (p PseudoClassVariant) base() string {
	return p.name
}

var pseudoClassVariants = []PseudoClassVariant{
	{"hover"},
	{"focus"},
	{"focus-within"},
	{"focus-visible"},
	{"active"},
	{"visited"},
	{"target"},
}

type PseudoElementVariant struct {
	name  string
	value string
}

func PseudoElementNameSameAsValue(n string) PseudoElementVariant {
	return PseudoElementVariant{name: n, value: n}
}

func (p PseudoElementVariant) convert(arbitaryValue string, slashText string, c CSS) []CSS {
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

func (m DoublePsuedoElementVariant) convert(arbitraryValue string, slashText string, c CSS) []CSS {
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

func (s customSupportsVariant) convert(arbitraryValue string, slashText string, c CSS) []CSS {
	c.SupportsStatements = append(c.SupportsStatements, s.value)
	return []CSS{c}
}
func (s customSupportsVariant) base() string {
	return "supports" + s.name
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

func (b BreakpointsVariant) convert(arbitraryValue string, slashText string, c CSS) []CSS {
	if arbitraryValue != "" {
		return nil
	}
	c.MediaQueries = append(c.MediaQueries, "min-width: "+b.value)
	return []CSS{c}
}
func (b BreakpointsVariant) base() string {
	return b.name
}

type groupVariant struct {
}

//////////////////////////////////////////// UTILS

func concatMaps[K comparable, V any](theMaps ...map[K]V) map[K]V {
	out := map[K]V{}
	for _, aMap := range theMaps {
		maps.Copy(out, aMap)
	}
	return out
}

func parseFraction(s string) (float64, error) {
	parts := strings.Split(s, "/")
	numerator, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return numerator, err
	}
	denominator, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return denominator, err
	}
	return numerator / denominator, nil
}

//////////////////////////////////////////// DEFAULTS

var defaultBreakpoints = map[string]string{
	"sm":  "640px",
	"md":  "768px",
	"lg":  "1024px",
	"xl":  "1280px",
	"2xl": "1536px",
}

var defaultNums = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12",
	"14", "16",
	"20", "24", "28", "32", "36", "40", "44", "48", "52", "56", "60", "64",
	"72", "80",
	"96",
}
var defaultFractions = []string{
	"1/2",
	"1/3", "2/3",
	"1/4", "2/4", "3/4",
	"1/5", "2/5", "3/5", "4/5",
	"1/6", "2/6", "3/6", "4/6", "5/6",
	"1/12", "2/12", "3/12", "4/12", "5/12", "6/12", "7/12", "8/12", "9/12", "10/12", "11/12",
}

// https://tailwindcss.com/docs/customizing-colors
var defaultColors = map[string]string{
	"slate-50":  "#f8fafc",
	"slate-100": "#f1f5f9",
	"slate-200": "#e2e8f0",
	"slate-300": "#cbd5e1",
	"slate-400": "#94a3b8",
	"slate-500": "#64748b",
	"slate-600": "#475569",
	"slate-700": "#334155",
	"slate-800": "#1e293b",
	"slate-900": "#0f172a",
	"slate-950": "#020617",

	"gray-50":  "#f9fafb",
	"gray-100": "#f3f4f6",
	"gray-200": "#e5e7eb",
	"gray-300": "#d1d5db",
	"gray-400": "#9ca3af",
	"gray-500": "#6b7280",
	"gray-600": "#4b5563",
	"gray-700": "#374151",
	"gray-800": "#1f2937",
	"gray-900": "#111827",
	"gray-950": "#030712",

	// I forgot zinc

	"neutral-50":  "#fafafa",
	"neutral-100": "#f5f5f5",
	"neutral-200": "#e5e5e5",
	"neutral-300": "#d4d4d4",
	"neutral-400": "#a3a3a3",
	"neutral-500": "#737373",
	"neutral-600": "#525252",
	"neutral-700": "#404040",
	"neutral-800": "#262626",
	"neutral-900": "#171717",
	"neutral-950": "#0a0a0a",

	"stone-50":  "#fafaf9",
	"stone-100": "#f5f5f4",
	"stone-200": "#e7e5e4",
	"stone-300": "#d6d3d1",
	"stone-400": "#a8a29e",
	"stone-500": "#78716c",
	"stone-600": "#57534e",
	"stone-700": "#44403c",
	"stone-800": "#292524",
	"stone-900": "#1c1917",
	"stone-950": "#0c0a09",

	"red-50":  "#fef2f2",
	"red-100": "#fee2e2",
	"red-200": "#fecaca",
	"red-300": "#fca5a5",
	"red-400": "#f87171",
	"red-500": "#ef4444",
	"red-600": "#dc2626",
	"red-700": "#b91c1c",
	"red-800": "#991b1b",
	"red-900": "#7f1d1d",
	"red-950": "#450a0a",

	"orange-50":  "#fff7ed",
	"orange-100": "#ffedd5",
	"orange-200": "#fed7aa",
	"orange-300": "#fdba74",
	"orange-400": "#fb923c",
	"orange-500": "#f97316",
	"orange-600": "#ea580c",
	"orange-700": "#c2410c",
	"orange-800": "#9a3412",
	"orange-900": "#7c2d12",
	"orange-950": "#431407",

	"amber-50":  "#fffbeb",
	"amber-100": "#fef3c7",
	"amber-200": "#fde68a",
	"amber-300": "#fcd34d",
	"amber-400": "#fbbf24",
	"amber-500": "#f59e0b",
	"amber-600": "#d97706",
	"amber-700": "#b45309",
	"amber-800": "#92400e",
	"amber-900": "#78350f",
	"amber-950": "#451a03",

	"yellow-50":  "#fefce8",
	"yellow-100": "#fef9c3",
	"yellow-200": "#fef08a",
	"yellow-300": "#fde047",
	"yellow-400": "#facc15",
	"yellow-500": "#eab308",
	"yellow-600": "#ca8a04",
	"yellow-700": "#a16207",
	"yellow-800": "#854d0e",
	"yellow-900": "#713f12",
	"yellow-950": "#422006",

	"lime-50":  "#f7fee7",
	"lime-100": "#ecfccb",
	"lime-200": "#d9f99d",
	"lime-300": "#bef264",
	"lime-400": "#a3e635",
	"lime-500": "#84cc16",
	"lime-600": "#65a30d",
	"lime-700": "#4d7c0f",
	"lime-800": "#3f6212",
	"lime-900": "#365314",
	"lime-950": "#1a2e05",

	"green-50":  "#f0fdf4",
	"green-100": "#dcfce7",
	"green-200": "#bbf7d0",
	"green-300": "#86efac",
	"green-400": "#4ade80",
	"green-500": "#22c55e",
	"green-600": "#16a34a",
	"green-700": "#15803d",
	"green-800": "#166534",
	"green-900": "#14532d",
	"green-950": "#052e16",

	"emerald-50":  "#ecfdf5",
	"emerald-100": "#d1fae5",
	"emerald-200": "#a7f3d0",
	"emerald-300": "#6ee7b7",
	"emerald-400": "#34d399",
	"emerald-500": "#10b981",
	"emerald-600": "#059669",
	"emerald-700": "#047857",
	"emerald-800": "#065f46",
	"emerald-900": "#064e3b",
	"emerald-950": "#022c22",

	"teal-50":  "#f0fdfa",
	"teal-100": "#ccfbf1",
	"teal-200": "#99f6e4",
	"teal-300": "#5eead4",
	"teal-400": "#2dd4bf",
	"teal-500": "#14b8a6",
	"teal-600": "#0d9488",
	"teal-700": "#0f766e",
	"teal-800": "#115e59",
	"teal-900": "#134e4a",
	"teal-950": "#042f2e",

	"cyan-50":  "#ecfeff",
	"cyan-100": "#cffafe",
	"cyan-200": "#a5f3fc",
	"cyan-300": "#67e8f9",
	"cyan-400": "#22d3ee",
	"cyan-500": "#06b6d4",
	"cyan-600": "#0891b2",
	"cyan-700": "#0e7490",
	"cyan-800": "#155e75",
	"cyan-900": "#164e63",
	"cyan-950": "#083344",

	"sky-50":  "#f0f9ff",
	"sky-100": "#e0f2fe",
	"sky-200": "#bae6fd",
	"sky-300": "#7dd3fc",
	"sky-400": "#38bdf8",
	"sky-500": "#0ea5e9",
	"sky-600": "#0284c7",
	"sky-700": "#0369a1",
	"sky-800": "#075985",
	"sky-900": "#0c4a6e",
	"sky-950": "#082f49",

	"blue-50":  "#eff6ff",
	"blue-100": "#dbeafe",
	"blue-200": "#bfdbfe",
	"blue-300": "#93c5fd",
	"blue-400": "#60a5fa",
	"blue-500": "#3b82f6",
	"blue-600": "#2563eb",
	"blue-700": "#1d4ed8",
	"blue-800": "#1e40af",
	"blue-900": "#1e3a8a",
	"blue-950": "#172554",

	"violet-50":  "#f5f3ff",
	"violet-100": "#ede9fe",
	"violet-200": "#ddd6fe",
	"violet-300": "#c4b5fd",
	"violet-400": "#a78bfa",
	"violet-500": "#8b5cf6",
	"violet-600": "#7c3aed",
	"violet-700": "#6d28d9",
	"violet-800": "#5b21b6",
	"violet-900": "#4c1d95",
	"violet-950": "#2e1065",

	"purple-50":  "#faf5ff",
	"purple-100": "#f3e8ff",
	"purple-200": "#e9d5ff",
	"purple-300": "#d8b4fe",
	"purple-400": "#c084fc",
	"purple-500": "#a855f7",
	"purple-600": "#9333ea",
	"purple-700": "#7e22ce",
	"purple-800": "#6b21a8",
	"purple-900": "#581c87",
	"purple-950": "#3b0764",

	"fuchsia-50":  "#fdf4ff",
	"fuchsia-100": "#fae8ff",
	"fuchsia-200": "#f5d0fe",
	"fuchsia-300": "#f0abfc",
	"fuchsia-400": "#e879f9",
	"fuchsia-500": "#d946ef",
	"fuchsia-600": "#c026d3",
	"fuchsia-700": "#a21caf",
	"fuchsia-800": "#86198f",
	"fuchsia-900": "#701a75",
	"fuchsia-950": "#4a044e",

	"pink-50":  "#fdf2f8",
	"pink-100": "#fce7f3",
	"pink-200": "#fbcfe8",
	"pink-300": "#f9a8d4",
	"pink-400": "#f472b6",
	"pink-500": "#ec4899",
	"pink-600": "#db2777",
	"pink-700": "#be185d",
	"pink-800": "#9d174d",
	"pink-900": "#831843",
	"pink-950": "#500724",

	"rose-50":  "#fff1f2",
	"rose-100": "#ffe4e6",
	"rose-200": "#fecdd3",
	"rose-300": "#fda4af",
	"rose-400": "#fb7185",
	"rose-500": "#f43f5e",
	"rose-600": "#e11d48",
	"rose-700": "#be123c",
	"rose-800": "#9f1239",
	"rose-900": "#881337",
	"rose-950": "#4c0519",
}
