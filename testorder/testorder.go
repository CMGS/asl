// Package testorder flags helper declarations placed above test funcs in _test.go files.
package testorder

import (
	"go/ast"
	"go/token"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

var (
	Analyzer = &analysis.Analyzer{
		Name: "testorder",
		Doc:  "flag helper declarations above test funcs in _test.go files",
		Run:  run,
	}

	testPrefixes = []string{"Test", "Benchmark", "Fuzz", "Example"}
)

func run(pass *analysis.Pass) (any, error) {
	for f := range source.Files(pass) {
		if source.IsTest(pass, f) {
			checkFile(pass, f)
		}
	}
	return nil, nil
}

func checkFile(pass *analysis.Pass, f *ast.File) {
	lastTest := token.NoPos
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && isTestFunc(fd) {
			lastTest = d.Pos()
		}
	}
	if !lastTest.IsValid() {
		return
	}
	for _, d := range f.Decls {
		if d.Pos() >= lastTest {
			return
		}
		switch dd := d.(type) {
		case *ast.FuncDecl:
			if !isTestFunc(dd) {
				pass.Reportf(dd.Pos(), "helper %s above test funcs; move it below the last test", dd.Name.Name)
			}
		case *ast.GenDecl:
			if dd.Tok == token.TYPE {
				pass.Reportf(dd.Pos(), "helper type above test funcs; move it below the last test")
			}
		}
	}
}

func isTestFunc(fd *ast.FuncDecl) bool {
	return fd.Recv == nil && slices.ContainsFunc(testPrefixes, func(p string) bool {
		return strings.HasPrefix(fd.Name.Name, p)
	})
}
