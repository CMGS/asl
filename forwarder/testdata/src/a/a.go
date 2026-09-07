package a

import "os"

type store struct{ dir string }

type closer interface {
	close() error
}

// present reports whether name exists in the store.
func (s *store) present(name string) bool { // want `single-use forwarder present spans 4 lines`
	return exists(s.dir + "/" + name)
}

func (s *store) close() error {
	return os.Remove(s.dir)
}

func (s *store) Has(name string) bool { return s.present(name) }

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func short() int { return other() }

func other() int { return 1 }

func twice() int {
	return 2
}

func Use() int { return short() + twice() + twice() }

func asValue() int {
	return 3
}

func Take() func() int { return asValue }

var _ closer = (*store)(nil)

func conditions() []int {
	return []int{
		1,
		2,
		3,
	}
}

func Conditions() []int { return conditions() }
