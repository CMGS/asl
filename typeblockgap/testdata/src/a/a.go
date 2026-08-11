package a

type catalogVolume struct { // want `type catalogVolume is split from its first method allowed by type resolvedVolume; keep the type declaration and its method set contiguous`
	tenants []string
}

type resolvedVolume struct {
	applied string
}

func (v catalogVolume) allowed(tenant string) bool { return len(v.tenants) == 0 }

type Session struct{} // want `type Session is split from its first method Close by method tracker.reset; keep the type declaration and its method set contiguous`

func (t tracker) reset() {}

func (s *Session) Close() error { return nil }

type tracker struct{}

func (t tracker) record() {}

type Registry struct{} // want `type Registry is split from its first method Lookup by standalone function normalize; keep the type declaration and its method set contiguous`

func normalize(name string) string { return name }

func (r *Registry) Lookup(name string) bool { return name == "" }
