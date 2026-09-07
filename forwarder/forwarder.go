// Package forwarder flags unexported one-statement funcs with a single caller whose inlining saves at least three lines and moves at most two.
package forwarder

import (
	"go/ast"
	"go/build/constraint"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

const (
	minLines     = 4
	maxBodyLines = 2
)

var Analyzer = &analysis.Analyzer{
	Name: "forwarder",
	Doc:  "flag single-use unexported forwarders of four or more lines; inline them or record them as kept",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if source.ShadowedByTestVariant(pass) {
		return nil, nil
	}
	refs, calls := countUses(pass)
	for f := range source.Files(pass) {
		if source.IsTest(pass, f) || buildConstrained(f) {
			continue
		}
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || !isForwarder(fd) {
				continue
			}
			obj := pass.TypesInfo.Defs[fd.Name]
			if obj == nil || refs[obj] != 1 || calls[obj] != 1 || implementsInterface(pass, obj) {
				continue
			}
			start := fd.Pos()
			if fd.Doc != nil {
				start = fd.Doc.Pos()
			}
			lines := pass.Fset.Position(fd.End()).Line - pass.Fset.Position(start).Line + 1
			stmt := fd.Body.List[0]
			if lines < minLines || pass.Fset.Position(stmt.End()).Line-pass.Fset.Position(stmt.Pos()).Line+1 > maxBodyLines {
				continue
			}
			pass.Reportf(fd.Pos(), "single-use forwarder %s spans %d lines; inline it at its one call site or record it as kept", fd.Name.Name, lines)
		}
	}
	return nil, nil
}

// countUses tallies every reference and every call-position reference of each object across the package.
func countUses(pass *analysis.Pass) (refs, calls map[types.Object]int) {
	refs, calls = map[types.Object]int{}, map[types.Object]int{}
	for _, f := range pass.Files {
		for n := range ast.Preorder(f) {
			switch n := n.(type) {
			case *ast.Ident:
				if obj := pass.TypesInfo.Uses[n]; obj != nil {
					refs[obj]++
				}
			case *ast.CallExpr:
				if id := callee(n.Fun); id != nil {
					if obj := pass.TypesInfo.Uses[id]; obj != nil {
						calls[obj]++
					}
				}
			}
		}
	}
	return refs, calls
}

// buildConstrained reports a //go:build file: its one-liners usually have a twin under another tag, so the caller cannot inline them.
func buildConstrained(f *ast.File) bool {
	for _, cg := range f.Comments {
		if cg.Pos() > f.Package {
			return false
		}
		for _, c := range cg.List {
			if constraint.IsGoBuild(c.Text) {
				return true
			}
		}
	}
	return false
}

func callee(e ast.Expr) *ast.Ident {
	switch e := e.(type) {
	case *ast.Ident:
		return e
	case *ast.SelectorExpr:
		return e.Sel
	case *ast.ParenExpr:
		return callee(e.X)
	case *ast.IndexExpr:
		return callee(e.X)
	case *ast.IndexListExpr:
		return callee(e.X)
	}
	return nil
}

func isForwarder(fd *ast.FuncDecl) bool {
	name := fd.Name.Name
	if ast.IsExported(name) || name == "main" || name == "init" || fd.Body == nil || len(fd.Body.List) != 1 {
		return false
	}
	switch s := fd.Body.List[0].(type) {
	case *ast.ReturnStmt:
		return len(s.Results) == 1
	case *ast.ExprStmt:
		_, isCall := s.X.(*ast.CallExpr)
		return isCall
	}
	return false
}

// implementsInterface reports whether obj is a method whose name and signature match a package-level interface's method.
func implementsInterface(pass *analysis.Pass, obj types.Object) bool {
	fn, ok := obj.(*types.Func)
	if !ok || fn.Signature().Recv() == nil {
		return false
	}
	scope := pass.Pkg.Scope()
	for _, name := range scope.Names() {
		tn, ok := scope.Lookup(name).(*types.TypeName)
		if !ok {
			continue
		}
		iface, ok := tn.Type().Underlying().(*types.Interface)
		if !ok {
			continue
		}
		for m := range iface.Methods() {
			if m.Name() == fn.Name() && types.Identical(m.Type(), fn.Type()) {
				return true
			}
		}
	}
	return false
}
