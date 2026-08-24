package a

type Size struct{}

func NewSize() *Size { return nil }

func (s *Size) Total() int { return 0 }

// Sanctioned adjacency: a result type directly above the method consuming it.
type SizeSpec struct{ n int }

func (s *Size) Spec() SizeSpec { return SizeSpec{} }

func (s *Size) Reset() {}

type Pty struct{}

func (p *Pty) Close() error { return nil }

// Producer trails the type it produces without splitting either method set.
func (s *Size) OpenPty() *Pty { return nil }

func lowHelper() int { return 0 }

type vocabulary struct{ kind string }

func trailing(v vocabulary) string { return v.kind }

type Meter struct{}

func (m *Meter) Read() {}

type (
	MeterSpec struct{}
	MeterOpts struct{}
)

func (m *Meter) Spec(o MeterOpts) MeterSpec { return MeterSpec{} }
