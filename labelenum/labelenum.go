// Package labelenum flags comments that enumerate metric label values in files that build Prometheus collectors.
package labelenum

import (
	"go/ast"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

const prometheusPath = "github.com/prometheus/client_golang/prometheus"

var (
	Analyzer = &analysis.Analyzer{
		Name: "labelenum",
		Doc:  "flag comments that enumerate label values (result=ok|failed) next to Prometheus collectors; values drift, tests and constants do not",
		Run:  run,
	}
	enumeration = regexp.MustCompile(`(?:^|[^\w-])([A-Za-z_]+)=([A-Za-z_]+(?:\|[A-Za-z_]+)+)`)
)

func run(pass *analysis.Pass) (any, error) {
	for f := range source.Files(pass) {
		if !importsPrometheus(f) {
			continue
		}
		for _, cg := range f.Comments {
			for _, c := range cg.List {
				if m := enumeration.FindStringSubmatch(c.Text); m != nil {
					pass.Reportf(c.Pos(), "comment enumerates the values of %s; name the meaning and let a test or constants own the values", m[1])
				}
			}
		}
	}
	return nil, nil
}

func importsPrometheus(f *ast.File) bool {
	for _, imp := range f.Imports {
		if path, err := strconv.Unquote(imp.Path.Value); err == nil && strings.HasPrefix(path, prometheusPath) {
			return true
		}
	}
	return false
}
