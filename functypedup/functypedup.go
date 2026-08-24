// Package functypedup flags a func type spelled out repeatedly instead of named once.
package functypedup

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

// minLen skips short callback shapes (func() error) where a named type buys nothing.
const minLen = 40

var Analyzer = &analysis.Analyzer{
	Name: "functypedup",
	Doc:  "flag func types spelling the same contract more than once in struct fields or results",
	Run:  run,
}

// site records one spelled-out func type: a struct field (by name) or a func result.
type site struct {
	pos    token.Pos
	field  string
	result bool
}

func run(pass *analysis.Pass) (any, error) {
	groups := map[string][]site{}
	for f := range source.Files(pass) {
		if source.IsTest(pass, f) {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch d := n.(type) {
			case *ast.StructType:
				for _, field := range d.Fields.List {
					if len(field.Names) > 0 {
						record(pass, groups, field.Type, site{field: field.Names[0].Name})
					}
				}
			case *ast.FuncDecl:
				if d.Type.Results != nil {
					for _, r := range d.Type.Results.List {
						record(pass, groups, r.Type, site{result: true})
					}
				}
			}
			return true
		})
	}
	for sig, sites := range groups {
		report(pass, sig, sites)
	}
	return nil, nil
}

// report flags a group only when it spells one contract: a func result (factory) is always a
// contract; same-shape fields count only under the same name — coincidental shape twins with
// different meanings stay inline.
func report(pass *analysis.Pass, sig string, sites []site) {
	if len(sites) < 2 {
		return
	}
	eligible := sites
	if !slices.ContainsFunc(sites, func(s site) bool { return s.result }) {
		names := map[string]int{}
		for _, s := range sites {
			names[s.field]++
		}
		eligible = slices.DeleteFunc(slices.Clone(sites), func(s site) bool { return names[s.field] < 2 })
		if len(eligible) < 2 {
			return
		}
	}
	for _, s := range eligible {
		pass.Reportf(s.pos, "func type %s spelled %d times; declare a named type", sig, len(eligible))
	}
}

func record(pass *analysis.Pass, groups map[string][]site, e ast.Expr, s site) {
	if _, ok := e.(*ast.FuncType); !ok {
		return
	}
	sig, ok := pass.TypesInfo.TypeOf(e).(*types.Signature)
	if !ok {
		return
	}
	if canon := canonical(sig); len(canon) >= minLen {
		s.pos = e.Pos()
		groups[canon] = append(groups[canon], s)
	}
}

// canonical prints the signature without parameter names so naming differences still dedupe.
func canonical(sig *types.Signature) string {
	qf := func(p *types.Package) string { return p.Name() }
	var b strings.Builder
	b.WriteString("func(")
	params := sig.Params()
	for i := range params.Len() {
		if i > 0 {
			b.WriteString(", ")
		}
		t := params.At(i).Type()
		if sig.Variadic() && i == params.Len()-1 {
			b.WriteString("...")
			t = t.(*types.Slice).Elem()
		}
		b.WriteString(types.TypeString(t, qf))
	}
	b.WriteString(")")
	results := sig.Results()
	switch results.Len() {
	case 0:
	case 1:
		b.WriteString(" " + types.TypeString(results.At(0).Type(), qf))
	default:
		b.WriteString(" (")
		for i := range results.Len() {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(types.TypeString(results.At(i).Type(), qf))
		}
		b.WriteString(")")
	}
	return b.String()
}
