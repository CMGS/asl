// Package source exposes what an analyzer inspects: the checkable files of a pass and the declarations of a file.
package source

import (
	"go/ast"
	"iter"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Files yields the hand-written files of pass — not cgo output or generated code.
func Files(pass *analysis.Pass) iter.Seq[*ast.File] {
	return func(yield func(*ast.File) bool) {
		for _, f := range pass.Files {
			if strings.HasSuffix(filename(pass, f), ".go") && !ast.IsGenerated(f) && !yield(f) {
				return
			}
		}
	}
}

// IsTest reports whether f is a _test.go file.
func IsTest(pass *analysis.Pass, f *ast.File) bool {
	return strings.HasSuffix(filename(pass, f), "_test.go")
}

func filename(pass *analysis.Pass, f *ast.File) string {
	return pass.Fset.Position(f.Pos()).Filename
}
