// Package typeblockgap flags declarations placed between a type declaration and that type's first method.
package typeblockgap

import (
	"go/ast"
	"slices"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

var Analyzer = &analysis.Analyzer{
	Name: "typeblockgap",
	Doc:  "flag declarations placed between a type declaration and that type's first method, or a type declared below its methods",
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
	first := firstMethods(seq)
	for owner, block := range seq {
		m := blockFirstMethod(block, first)
		if m >= 0 && m < owner {
			pass.Reportf(block.Node.Pos(), "type %s declared below its method %s; declare the type above its method set", block.Name, seq[m].Name)
			continue
		}
		for i := owner + 1; i < m; i++ {
			if exempt(pass, seq, i, owner, m) {
				continue
			}
			pass.Reportf(block.Node.Pos(), "type %s is split from its first method %s by %s; keep the type declaration and its method set contiguous", block.Name, seq[m].Name, seq[i])
			break
		}
	}
}

func firstMethods(seq []source.Decl) map[string]int {
	first := map[string]int{}
	for i, d := range seq {
		if _, seen := first[d.Recv]; d.Recv != "" && !seen {
			first[d.Recv] = i
		}
	}
	return first
}

// blockFirstMethod treats a grouped declaration as one block: its members share the earliest method of any member.
func blockFirstMethod(d source.Decl, first map[string]int) int {
	m := -1
	for _, t := range d.Types {
		if i, ok := first[t]; ok && (m < 0 || i < m) {
			m = i
		}
	}
	return m
}

// exempt reports the sanctioned occupants: a producer of the owner, or a result type directly above the owner method consuming it.
func exempt(pass *analysis.Pass, seq []source.Decl, i, owner, m int) bool {
	d := seq[i]
	switch {
	case d.Types != nil:
		return i+1 == m && slices.ContainsFunc(d.Types, seq[m].Mentions)
	case d.Recv != "":
		return false
	default:
		return d.ProducesAny(pass, seq[owner:owner+1])
	}
}
