package source

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var testPrefixes = []string{"Test", "Benchmark", "Fuzz", "Example"}

// Decl is one top-level func, method, or type block of a file.
type Decl struct {
	Node  ast.Node
	Name  string
	Recv  string   // receiver type; "" for standalone funcs and type blocks
	Types []string // type names declared by a type block; nil for funcs
}

// IsFunc reports whether d is a standalone func.
func (d Decl) IsFunc() bool {
	return d.Recv == "" && d.Types == nil
}

// IsTest reports whether d is a test, benchmark, fuzz, or example func.
func (d Decl) IsTest() bool {
	return d.IsFunc() && slices.ContainsFunc(testPrefixes, func(p string) bool { return strings.HasPrefix(d.Name, p) })
}

// Mentions reports whether the func signature names t.
func (d Decl) Mentions(t string) bool {
	fd, ok := d.Node.(*ast.FuncDecl)
	if !ok {
		return false
	}
	for n := range ast.Preorder(fd.Type) {
		if id, ok := n.(*ast.Ident); ok && id.Name == t {
			return true
		}
	}
	return false
}

// Produces reports whether the func returns t or an interface t implements.
func (d Decl) Produces(pass *analysis.Pass, t string) bool {
	fd, ok := d.Node.(*ast.FuncDecl)
	if !ok || fd.Type.Results == nil {
		return false
	}
	obj := pass.Pkg.Scope().Lookup(t)
	return slices.ContainsFunc(fd.Type.Results.List, func(r *ast.Field) bool {
		if Ident(r.Type) == t {
			return true
		}
		iface, ok := pass.TypesInfo.TypeOf(r.Type).Underlying().(*types.Interface)
		return ok && obj != nil && (types.Implements(obj.Type(), iface) || types.Implements(types.NewPointer(obj.Type()), iface))
	})
}

// Decls lists the funcs, methods, and type blocks of f in source order.
func Decls(f *ast.File) []Decl {
	var seq []Decl
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			seq = append(seq, Decl{Node: d, Name: d.Name.Name, Recv: receiver(d)})
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			var block Decl
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if block.Node == nil {
					block.Node, block.Name = ts, ts.Name.Name
				}
				block.Types = append(block.Types, ts.Name.Name)
			}
			if block.Node != nil {
				seq = append(seq, block)
			}
		}
	}
	return seq
}

// Ident unwraps parenthesis, pointer, and type-parameter wrappers down to the type identifier.
func Ident(e ast.Expr) string {
	if paren, ok := e.(*ast.ParenExpr); ok {
		e = paren.X
	}
	if star, ok := e.(*ast.StarExpr); ok {
		e = star.X
	}
	switch x := e.(type) {
	case *ast.IndexExpr:
		e = x.X
	case *ast.IndexListExpr:
		e = x.X
	}
	if id, ok := e.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

func receiver(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return ""
	}
	return Ident(fd.Recv.List[0].Type)
}
