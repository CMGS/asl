// Package methodinterleave flags declarations of another owner placed between two methods of the same receiver.
package methodinterleave

import (
	"go/ast"
	"slices"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

var Analyzer = &analysis.Analyzer{
	Name: "methodinterleave",
	Doc:  "flag standalone funcs, types, and another receiver's methods declared between two methods of the same receiver",
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
	methods := map[string][]int{}
	var receivers []string
	for i, d := range seq {
		if d.Recv == "" {
			checkStandalone(pass, seq, i)
			continue
		}
		if _, seen := methods[d.Recv]; !seen {
			receivers = append(receivers, d.Recv)
		}
		methods[d.Recv] = append(methods[d.Recv], i)
	}
	for _, r := range receivers {
		checkResume(pass, seq, r, methods[r])
	}
}

func checkStandalone(pass *analysis.Pass, seq []source.Decl, i int) {
	d := seq[i]
	prev := nearestReceiver(seq, i, -1)
	next := nearestReceiver(seq, i, +1)
	if prev == "" || prev != next {
		return
	}
	// a type directly above a method naming it is the sanctioned adjacency (SizeSpec above Size.Spec)
	if d.Types != nil && seq[i+1].Recv != "" && slices.ContainsFunc(d.Types, seq[i+1].Mentions) {
		return
	}
	pass.Reportf(d.Node.Pos(), "%s declared between %s methods; keep the method set contiguous and move it above or below", d, prev)
}

// checkResume accepts a foreign block inside r's method set only while every later r method produces a type declared after the split.
func checkResume(pass *analysis.Pass, seq []source.Decl, r string, methods []int) {
	for k := 0; k+1 < len(methods); k++ {
		a, b := methods[k], methods[k+1]
		if !slices.ContainsFunc(seq[a+1:b], func(d source.Decl) bool { return d.Recv != "" && d.Recv != r }) {
			continue
		}
		split := a + 1 + slices.IndexFunc(seq[a+1:b], func(d source.Decl) bool { return d.Types != nil || d.Recv != "" })
		for _, m := range methods[k+1:] {
			if !seq[m].ProducesAny(pass, seq[a+1:m]) {
				pass.Reportf(seq[m].Node.Pos(), "%s resumes the %s method set after %s; keep the method set contiguous, only producers trail a foreign type block", seq[m], r, seq[split])
				return
			}
		}
		return
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
