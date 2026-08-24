// Package methodpartition flags unexported methods declared above an exported method of the same receiver in the same file.
package methodpartition

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

var Analyzer = &analysis.Analyzer{
	Name: "methodpartition",
	Doc:  "flag unexported methods declared above an exported method of the same receiver",
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
	last := map[string]source.Decl{}
	for _, d := range seq {
		if d.Recv != "" && ast.IsExported(d.Name) {
			last[d.Recv] = d
		}
	}
	for _, d := range seq {
		if d.Recv == "" || ast.IsExported(d.Name) {
			continue
		}
		if e, ok := last[d.Recv]; ok && d.Node.Pos() < e.Node.Pos() {
			pass.Reportf(d.Node.Pos(), "unexported method %s.%s declared above exported method %s; move unexported methods below the exported set", d.Recv, d.Name, e.Name)
		}
	}
}
