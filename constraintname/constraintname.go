// Package constraintname flags type-parameter constraints that are not named types.
package constraintname

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

var Analyzer = &analysis.Analyzer{
	Name: "constraintname",
	Doc:  "flag inline interface, union, and other unnamed constraints in type-parameter lists",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for f := range source.Files(pass) {
		ast.Inspect(f, func(n ast.Node) bool {
			var params *ast.FieldList
			switch d := n.(type) {
			case *ast.FuncDecl:
				params = d.Type.TypeParams
			case *ast.TypeSpec:
				params = d.TypeParams
			default:
				return true
			}
			if params == nil {
				return true
			}
			for _, field := range params.List {
				if !isNamed(field.Type) {
					pass.Reportf(field.Type.Pos(), "inline interface constraint; declare a named constraint type")
				}
			}
			return true
		})
	}
	return nil, nil
}

func isNamed(e ast.Expr) bool {
	switch x := e.(type) {
	case *ast.IndexExpr:
		e = x.X
	case *ast.IndexListExpr:
		e = x.X
	}
	switch e.(type) {
	case *ast.Ident, *ast.SelectorExpr:
		return true
	}
	return false
}
