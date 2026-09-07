package a

import "time"

type conf struct {
	name  string
	ptr   *int
	items []string
	n     uint
	d     time.Duration
}

func (c conf) label(def string) string {
	if c.name != "" { // want `if/return fallback on c.name is cmp.Or\(c.name, def\)`
		return c.name
	}
	return def
}

func pick(a, b *int) *int {
	if a == nil { // want `if/return fallback on a is cmp.Or\(a, b\)`
		return b
	}
	return a
}

func assign(c *conf, def string) {
	if c.name == "" { // want `zero-value fallback on c.name is c.name = cmp.Or\(c.name, def\)`
		c.name = def
	}
}

func unsigned(c conf) uint {
	if c.n > 0 { // want `if/return fallback on c.n is cmp.Or\(c.n, 4\)`
		return c.n
	}
	return 4
}

func signedStaysHuman(c conf) time.Duration {
	if c.d > 0 {
		return c.d
	}
	return time.Second
}

func sliceIsNotComparable(c conf) []string {
	if c.items != nil {
		return c.items
	}
	return []string{"x"}
}

func differentResults(c conf, def string) string {
	if c.name != "" {
		return c.name + "!"
	}
	return def
}

func withElse(c conf, def string) string {
	if c.name != "" {
		return c.name
	} else {
		return def
	}
}

func lazyFallbackStays(err error) error {
	if err != nil {
		return err
	}
	return finish()
}

func finish() error { return nil }

func assignCallStays(c *conf) {
	if c.name == "" {
		c.name = pickName()
	}
}

func pickName() string { return "x" }
