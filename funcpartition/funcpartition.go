// Package funcpartition flags unexported standalone functions declared above an exported standalone function in the same file.
package funcpartition

import (
	"go/ast"
	"go/token"
	"go/types"
	"maps"
	"slices"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

var Analyzer = &analysis.Analyzer{
	Name: "funcpartition",
	Doc:  "flag unexported standalone functions declared above an exported standalone function",
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

func checkFile(pass *analysis.Pass, f *ast.File) {
	fileTypes := map[string]bool{}
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok {
				fileTypes[ts.Name.Name] = true
			}
		}
	}
	var lastExported *ast.FuncDecl
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv == nil && fd.Name.IsExported() {
			lastExported = fd
		}
	}
	if lastExported == nil {
		return
	}
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Recv != nil || fd.Name.IsExported() || fd.Pos() >= lastExported.Pos() {
			continue
		}
		if fd.Name.Name == "main" || fd.Name.Name == "init" || constructs(pass, fd, fileTypes) {
			continue
		}
		pass.Reportf(fd.Pos(), "unexported function %s declared above exported function %s; move unexported helpers below the exported set", fd.Name.Name, lastExported.Name.Name)
	}
}

// constructs reports whether fd returns a type declared in the same file, or an interface one implements — the producer shape that belongs with its type block.
func constructs(pass *analysis.Pass, fd *ast.FuncDecl, fileTypes map[string]bool) bool {
	if fd.Type.Results == nil {
		return false
	}
	for _, r := range fd.Type.Results.List {
		t := r.Type
		if star, ok := t.(*ast.StarExpr); ok {
			t = star.X
		}
		switch x := t.(type) {
		case *ast.IndexExpr:
			t = x.X
		case *ast.IndexListExpr:
			t = x.X
		}
		if id, ok := t.(*ast.Ident); ok && fileTypes[id.Name] {
			return true
		}
		if iface, ok := pass.TypesInfo.TypeOf(r.Type).Underlying().(*types.Interface); ok && slices.ContainsFunc(slices.Collect(maps.Keys(fileTypes)), func(name string) bool { return implements(pass, name, iface) }) {
			return true
		}
	}
	return false
}

func implements(pass *analysis.Pass, name string, iface *types.Interface) bool {
	obj := pass.Pkg.Scope().Lookup(name)
	return obj != nil && (types.Implements(obj.Type(), iface) || types.Implements(types.NewPointer(obj.Type()), iface))
}
