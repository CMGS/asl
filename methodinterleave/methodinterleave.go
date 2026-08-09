// Package methodinterleave flags standalone funcs and types declared between two methods of the same receiver.
package methodinterleave

import (
	"go/ast"
	"go/token"

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
	node ast.Node
	name string
	recv string // receiver type; "" for standalone funcs and types
	typ  bool
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
			for _, spec := range d.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					seq = append(seq, decl{node: ts, name: ts.Name.Name, typ: true})
				}
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
		if d.typ && consumedByNext(seq, i) {
			continue
		}
		what := "standalone function"
		if d.typ {
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

// consumedByNext reports the sanctioned adjacency: a method-less type directly
// above a method that names it in its signature (`SizeSpec` above `Size.Spec`).
func consumedByNext(seq []decl, i int) bool {
	if i+1 >= len(seq) || seq[i+1].recv == "" {
		return false
	}
	fd, ok := seq[i+1].node.(*ast.FuncDecl)
	if !ok {
		return false
	}
	found := false
	ast.Inspect(fd.Type, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == seq[i].name {
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
