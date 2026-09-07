// Package functypedup flags a func type spelled out repeatedly instead of named once.
package functypedup

import (
	"go/ast"
	"go/token"
	"go/types"
	"maps"
	"slices"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

// minLen skips short callback shapes (func() error) where a named type buys nothing.
const minLen = 40

var Analyzer = &analysis.Analyzer{
	Name: "functypedup",
	Doc:  "flag func types spelling the same contract more than once in struct fields, parameters, package vars, or results",
	Run:  run,
}

// site records one spelled-out func type: a named field, parameter, or var, or a func result.
type site struct {
	pos    token.Pos
	name   string
	result bool
}

func run(pass *analysis.Pass) (any, error) {
	groups := map[string][]site{}
	for f := range source.Files(pass) {
		if source.IsTest(pass, f) {
			continue
		}
		for _, d := range f.Decls {
			if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.VAR {
				for _, spec := range gd.Specs {
					if vs := spec.(*ast.ValueSpec); vs.Type != nil {
						record(pass, groups, vs.Type, site{name: vs.Names[0].Name})
					}
				}
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch d := n.(type) {
			case *ast.StructType:
				for _, field := range d.Fields.List {
					if len(field.Names) > 0 {
						record(pass, groups, field.Type, site{name: field.Names[0].Name})
					}
				}
			case *ast.FuncDecl:
				for _, p := range d.Type.Params.List {
					if len(p.Names) > 0 {
						record(pass, groups, p.Type, site{name: p.Names[0].Name})
					}
				}
				if d.Type.Results != nil {
					for _, r := range d.Type.Results.List {
						record(pass, groups, r.Type, site{result: true})
					}
				}
			}
			return true
		})
	}
	for _, sig := range slices.Sorted(maps.Keys(groups)) {
		report(pass, sig, groups[sig])
	}
	return nil, nil
}

// report flags a group only when it spells one contract: any func result, or one name at two sites.
func report(pass *analysis.Pass, sig string, sites []site) {
	if len(sites) < 2 {
		return
	}
	eligible := sites
	if !slices.ContainsFunc(sites, func(s site) bool { return s.result }) {
		names := map[string]int{}
		for _, s := range sites {
			names[s.name]++
		}
		eligible = slices.DeleteFunc(sites, func(s site) bool { return names[s.name] < 2 })
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
	// parameter names are dropped so naming differences still dedupe.
	canon := types.TypeString(types.NewSignatureType(nil, nil, nil, unnamed(sig.Params()), unnamed(sig.Results()), sig.Variadic()), (*types.Package).Name)
	if len(canon) >= minLen {
		s.pos = e.Pos()
		groups[canon] = append(groups[canon], s)
	}
}

func unnamed(t *types.Tuple) *types.Tuple {
	vars := make([]*types.Var, 0, t.Len())
	for v := range t.Variables() {
		vars = append(vars, types.NewVar(token.NoPos, nil, "", v.Type()))
	}
	return types.NewTuple(vars...)
}
