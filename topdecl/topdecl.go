// Package topdecl flags const/var declarations outside the single top-of-file blocks.
package topdecl

import (
	"go/ast"
	"go/token"
	"slices"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

var Analyzer = &analysis.Analyzer{
	Name: "topdecl",
	Doc:  "flag const/var declarations below the first func, split across blocks, or out of const-then-var order",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for f := range source.Files(pass) {
		checkFile(pass, f)
	}
	return nil, nil
}

func checkFile(pass *analysis.Pass, f *ast.File) {
	firstFunc := token.NoPos
	if i := slices.IndexFunc(f.Decls, func(d ast.Decl) bool { _, ok := d.(*ast.FuncDecl); return ok }); i >= 0 {
		firstFunc = f.Decls[i].Pos()
	}
	blocks := map[token.Token]int{}
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok == token.IMPORT || gd.Tok == token.TYPE || isCheckOnly(gd) {
			continue
		}
		if firstFunc.IsValid() && gd.Pos() > firstFunc {
			pass.Reportf(gd.Pos(), "%s declaration below the first func; move it into the top block", gd.Tok)
			continue
		}
		blocks[gd.Tok]++
		if blocks[gd.Tok] > 1 {
			pass.Reportf(gd.Pos(), "more than one top-level %s block; merge into a single block", gd.Tok)
		}
		if gd.Tok == token.CONST && blocks[token.VAR] > 0 {
			pass.Reportf(gd.Pos(), "const block below the var block; declare const first")
		}
	}
}

// isCheckOnly exempts compile-time interface checks (var _ Iface = ...), which stand alone by rule.
func isCheckOnly(gd *ast.GenDecl) bool {
	if gd.Tok != token.VAR {
		return false
	}
	for _, spec := range gd.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok || vs.Type == nil {
			return false
		}
		for _, name := range vs.Names {
			if name.Name != "_" {
				return false
			}
		}
	}
	return true
}
