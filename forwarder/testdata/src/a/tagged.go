//go:build ignore_never_set

package a

func (s *store) reset() error {
	return exists(s.dir) == false && false
}

func Reset(s *store) error { return s.reset() }
