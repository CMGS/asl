// Package source filters which files an analyzer may report on.
package source

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Checkable reports whether f is hand-written source — not cgo output or generated code.
func Checkable(pass *analysis.Pass, f *ast.File) bool {
	return strings.HasSuffix(pass.Fset.Position(f.Pos()).Filename, ".go") && !ast.IsGenerated(f)
}
