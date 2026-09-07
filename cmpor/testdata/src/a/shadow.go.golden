package a

import "github.com/google/go-cmp/cmp"

func shadowed(want, got string) string {
	if got != "" {
		return got
	}
	return want + cmp.Diff(want, got)
}
