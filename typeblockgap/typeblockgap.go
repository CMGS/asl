// Package typeblockgap flags declarations placed between a type declaration and that type's first method.
package typeblockgap

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

var Analyzer = &analysis.Analyzer{
	Name: "typeblockgap",
	Doc:  "flag declarations placed between a type declaration and that type's first method",
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
	seq := declSeq(f)
	first := firstMethods(seq)
	for owner, block := range seq {
		m := blockFirstMethod(block, first)
		for i := owner + 1; i < m; i++ {
			if exempt(pass, seq, i, owner, m) {
				continue
			}
			pass.Reportf(block.node.Pos(), "type %s is split from its first method %s by %s; keep the type declaration and its method set contiguous", block.name, seq[m].name, describe(seq[i]))
			break
		}
	}
}

func declSeq(f *ast.File) []decl {
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
	return seq
}

func firstMethods(seq []decl) map[string]int {
	first := map[string]int{}
	for i, d := range seq {
		if _, seen := first[d.recv]; d.recv != "" && !seen {
			first[d.recv] = i
		}
	}
	return first
}

// blockFirstMethod returns where the block's method set starts, or -1; a grouped
// declaration is one block, so its members share the earliest method of any member.
func blockFirstMethod(d decl, first map[string]int) int {
	m := -1
	for _, t := range d.types {
		if i, ok := first[t]; ok && (m < 0 || i < m) {
			m = i
		}
	}
	return m
}

// exempt reports the two sanctioned shapes: a producer of the owner type, and a
// result type directly above the owner method consuming it (`SizeSpec` above `Size.Spec`).
func exempt(pass *analysis.Pass, seq []decl, i, owner, m int) bool {
	d, o := seq[i], seq[owner]
	switch {
	case d.types != nil:
		return i+1 == m && slices.ContainsFunc(d.types, func(t string) bool { return consumedBy(seq[m], t) })
	case d.recv != "":
		return false
	default:
		return slices.ContainsFunc(o.types, func(t string) bool { return produces(pass, d, t) })
	}
}

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

// produces reports whether the func returns name or an interface it implements.
func produces(pass *analysis.Pass, d decl, name string) bool {
	fd, ok := d.node.(*ast.FuncDecl)
	if !ok || fd.Type.Results == nil {
		return false
	}
	obj := pass.Pkg.Scope().Lookup(name)
	return slices.ContainsFunc(fd.Type.Results.List, func(r *ast.Field) bool {
		if identName(r.Type) == name {
			return true
		}
		iface, ok := pass.TypesInfo.TypeOf(r.Type).Underlying().(*types.Interface)
		return ok && obj != nil && (types.Implements(obj.Type(), iface) || types.Implements(types.NewPointer(obj.Type()), iface))
	})
}

func describe(d decl) string {
	switch {
	case d.types != nil:
		return "type " + d.name
	case d.recv != "":
		return "method " + d.recv + "." + d.name
	default:
		return "standalone function " + d.name
	}
}

func receiverName(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return ""
	}
	return identName(fd.Recv.List[0].Type)
}

// identName unwraps pointer and type-parameter wrappers down to the type identifier.
func identName(t ast.Expr) string {
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
