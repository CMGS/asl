// Package methodpartition flags unexported methods declared above an exported method of the same receiver in the same file.
package methodpartition

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

var Analyzer = &analysis.Analyzer{
	Name: "methodpartition",
	Doc:  "flag unexported methods declared above an exported method of the same receiver",
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
	type exported struct {
		pos  token.Pos
		name string
	}
	last := map[string]exported{}
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Recv == nil || !fd.Name.IsExported() {
			continue
		}
		if recv := receiverName(fd); recv != "" {
			last[recv] = exported{pos: fd.Pos(), name: fd.Name.Name}
		}
	}
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Recv == nil || fd.Name.IsExported() {
			continue
		}
		recv := receiverName(fd)
		if e, ok := last[recv]; ok && fd.Pos() < e.pos {
			pass.Reportf(fd.Pos(), "unexported method %s.%s declared above exported method %s; move unexported methods below the exported set", recv, fd.Name.Name, e.name)
		}
	}
}

// receiverName unwraps pointer and type-parameter wrappers down to the receiver's type identifier.
func receiverName(fd *ast.FuncDecl) string {
	if len(fd.Recv.List) == 0 {
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
