// Package testorder flags helper declarations placed above test funcs in _test.go files.
package testorder

import (
	"go/ast"
	"slices"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

var Analyzer = &analysis.Analyzer{
	Name: "testorder",
	Doc:  "flag helper declarations above test funcs in _test.go files",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for f := range source.Files(pass) {
		if source.IsTest(pass, f) {
			checkFile(pass, f)
		}
	}
	return nil, nil
}

func checkFile(pass *analysis.Pass, f *ast.File) {
	seq := source.Decls(f)
	last := len(seq) - 1
	for last >= 0 && !seq[last].IsTest() {
		last--
	}
	for _, d := range slices.DeleteFunc(seq[:max(last, 0)], source.Decl.IsTest) {
		if d.Types != nil {
			pass.Reportf(d.Node.Pos(), "helper type above test funcs; move it below the last test")
		} else {
			pass.Reportf(d.Node.Pos(), "helper %s above test funcs; move it below the last test", d.Name)
		}
	}
}
