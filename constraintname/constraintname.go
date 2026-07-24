// Package constraintname flags inline interface literals used as type-parameter constraints.
package constraintname

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

var Analyzer = &analysis.Analyzer{
	Name: "constraintname",
	Doc:  "flag inline interface constraints in type-parameter lists",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, f := range pass.Files {
		if !source.Checkable(pass, f) {
			continue
		}
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
				if hasInterfaceLit(field.Type) {
					pass.Reportf(field.Type.Pos(), "inline interface constraint; declare a named constraint type")
				}
			}
			return true
		})
	}
	return nil, nil
}

func hasInterfaceLit(e ast.Expr) bool {
	found := false
	ast.Inspect(e, func(n ast.Node) bool {
		if _, ok := n.(*ast.InterfaceType); ok {
			found = true
		}
		return !found
	})
	return found
}
