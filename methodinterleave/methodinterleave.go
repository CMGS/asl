// Package methodinterleave flags standalone funcs and types declared between two methods of the same receiver.
package methodinterleave

import (
	"go/ast"
	"go/token"
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
	for _, f := range pass.Files {
		if source.Checkable(pass, f) {
			checkFile(pass, f)
		}
	}
	return nil, nil
}

type decl struct {
	node  ast.Node
	name  string
	recv  string   // receiver type; "" for standalone funcs and type blocks
	types []string // type names declared by a type block; nil for funcs
}

func checkFile(pass *analysis.Pass, f *ast.File) {
	var seq []decl
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			seq = append(seq, decl{node: d, name: d.Name.Name, recv: receiverName(d)})
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			var block decl
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if block.node == nil {
					block.node, block.name = ts, ts.Name.Name
				}
				block.types = append(block.types, ts.Name.Name)
			}
			if block.node != nil {
				seq = append(seq, block)
			}
		}
	}
	for i, d := range seq {
		if d.recv != "" {
			continue
		}
		prev := nearestReceiver(seq, i, -1)
		next := nearestReceiver(seq, i, +1)
		if prev == "" || prev != next {
			continue
		}
		if d.types != nil && seq[i+1].recv != "" && slices.ContainsFunc(d.types, func(t string) bool { return consumedBy(seq[i+1], t) }) {
			continue
		}
		what := "standalone function"
		if d.types != nil {
			what = "type"
		}
		pass.Reportf(d.node.Pos(), "%s %s declared between %s methods; keep the method set contiguous and move it above or below", what, d.name, prev)
	}
}

func nearestReceiver(seq []decl, i, step int) string {
	for j := i + step; j >= 0 && j < len(seq); j += step {
		if seq[j].recv != "" {
			return seq[j].recv
		}
	}
	return ""
}

// consumedBy reports the sanctioned adjacency: a type directly above a method naming it (SizeSpec above Size.Spec).
func consumedBy(d decl, name string) bool {
	fd, ok := d.node.(*ast.FuncDecl)
	if !ok {
		return false
	}
	found := false
	ast.Inspect(fd.Type, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == name {
			found = true
		}
		return !found
	})
	return found
}

func receiverName(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return ""
	}
	t := fd.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	switch x := t.(type) {
	case *ast.IndexExpr:
		t = x.X
	case *ast.IndexListExpr:
		t = x.X
	}
	if id, ok := t.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}
