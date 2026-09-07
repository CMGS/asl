package source

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ShadowedByTestVariant reports whether pass is the test-free variant of a package that has _test.go files, so a
// package-wide count taken here would miss the test callers the test variant sees.
func ShadowedByTestVariant(pass *analysis.Pass) bool {
	if len(pass.Files) == 0 {
		return false
	}
	for _, f := range pass.Files {
		if IsTest(pass, f) {
			return false
		}
	}
	entries, err := os.ReadDir(filepath.Dir(filename(pass, pass.Files[0])))
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), "_test.go") {
			return true
		}
	}
	return false
}
