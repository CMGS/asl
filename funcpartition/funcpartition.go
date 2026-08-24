// Package funcpartition flags unexported standalone functions declared above an exported standalone function in the same file.
package funcpartition

import (
	"go/ast"
	"slices"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

var Analyzer = &analysis.Analyzer{
	Name: "funcpartition",
	Doc:  "flag unexported standalone functions declared above an exported standalone function",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for f := range source.Files(pass) {
		checkFile(pass, f)
	}
	return nil, nil
}

func checkFile(pass *analysis.Pass, f *ast.File) {
	seq := source.Decls(f)
	var fileTypes []string
	last := -1
	for i, d := range seq {
		fileTypes = append(fileTypes, d.Types...)
		if d.IsFunc() && !d.IsTest() && ast.IsExported(d.Name) {
			last = i
		}
	}
	if last < 0 {
		return
	}
	for _, d := range seq[:last] {
		if !d.IsFunc() || d.IsTest() || ast.IsExported(d.Name) || d.Name == "main" || d.Name == "init" {
			continue
		}
		if slices.ContainsFunc(fileTypes, func(t string) bool { return d.Produces(pass, t) }) {
			continue
		}
		pass.Reportf(d.Node.Pos(), "unexported function %s declared above exported function %s; move unexported helpers below the exported set", d.Name, seq[last].Name)
	}
}
