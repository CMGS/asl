// Package methodinterleave flags standalone funcs and types declared between two methods of the same receiver.
package methodinterleave

import (
	"go/ast"
	"slices"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

var Analyzer = &analysis.Analyzer{
	Name: "methodinterleave",
	Doc:  "flag standalone funcs and types declared between two methods of the same receiver",
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
	for i, d := range seq {
		if d.Recv != "" {
			continue
		}
		prev := nearestReceiver(seq, i, -1)
		next := nearestReceiver(seq, i, +1)
		if prev == "" || prev != next {
			continue
		}
		// a type directly above a method naming it is the sanctioned adjacency (SizeSpec above Size.Spec)
		if d.Types != nil && seq[i+1].Recv != "" && slices.ContainsFunc(d.Types, seq[i+1].Mentions) {
			continue
		}
		what := "standalone function"
		if d.Types != nil {
			what = "type"
		}
		pass.Reportf(d.Node.Pos(), "%s %s declared between %s methods; keep the method set contiguous and move it above or below", what, d.Name, prev)
	}
}

func nearestReceiver(seq []source.Decl, i, step int) string {
	for j := i + step; j >= 0 && j < len(seq); j += step {
		if seq[j].Recv != "" {
			return seq[j].Recv
		}
	}
	return ""
}
