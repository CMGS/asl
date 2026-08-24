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
	Doc:  "flag const/var declarations below the first func, split across blocks, or out of const-then-var order, and interface checks away from their type",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for f := range source.Files(pass) {
		checkFile(pass, f)
	}
	return nil, nil
}

func checkFile(pass *analysis.Pass, f *ast.File) {
	checkPlacement(pass, f)
	firstFunc := token.NoPos
	if i := slices.IndexFunc(f.Decls, func(d ast.Decl) bool { _, ok := d.(*ast.FuncDecl); return ok }); i >= 0 {
		firstFunc = f.Decls[i].Pos()
	}
	blocks := map[token.Token]int{}
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok == token.IMPORT || gd.Tok == token.TYPE || isCheckOnly(d) {
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

// checkPlacement requires a check of a file-local type to sit immediately above that type, other checks between allowed.
func checkPlacement(pass *analysis.Pass, f *ast.File) {
	typeAt := map[string]int{}
	for i, d := range f.Decls {
		if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.TYPE {
			for _, spec := range gd.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					typeAt[ts.Name.Name] = i
				}
			}
		}
	}
	for i, d := range f.Decls {
		if !isCheckOnly(d) {
			continue
		}
		next := i + 1
		for next < len(f.Decls) && isCheckOnly(f.Decls[next]) {
			next++
		}
		for _, spec := range d.(*ast.GenDecl).Specs {
			for _, v := range spec.(*ast.ValueSpec).Values {
				if at, ok := typeAt[checkedType(v)]; ok && at != next {
					pass.Reportf(v.Pos(), "interface check for %s away from its type; place it immediately above the type", checkedType(v))
				}
			}
		}
	}
}

// checkedType names the type behind (*T)(nil), T{}, &T{}, or T(x).
func checkedType(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.CallExpr:
		return source.Ident(x.Fun)
	case *ast.CompositeLit:
		return source.Ident(x.Type)
	case *ast.UnaryExpr:
		return checkedType(x.X)
	}
	return ""
}

// isCheckOnly exempts compile-time interface checks (var _ Iface = ...), which stand alone by rule.
func isCheckOnly(d ast.Decl) bool {
	gd, ok := d.(*ast.GenDecl)
	if !ok || gd.Tok != token.VAR {
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
