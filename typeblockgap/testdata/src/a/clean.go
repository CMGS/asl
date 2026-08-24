package a

import "b"

type closer interface{ Close() error }

var _ closer = (*Handle)(nil)

type Handle struct{}

func (h *Handle) Close() error { return nil }

type Size struct{}

func NewSize() *Size { return &Size{} }

func (s *Size) Total() int { return 0 }

type Report struct{}

// Sanctioned adjacency: a result type directly above the owner method consuming it.
type ReportSpec struct{ n int }

func (r *Report) Spec() ReportSpec { return ReportSpec{} }

func (r *Report) Total() int { return 0 }

type (
	kind   string
	region string
)

type payload struct{ body []byte }

// A grouped declaration is one block: its members never split each other.
type (
	reader struct{}
	writer struct{}
)

func (r reader) read() int { return 0 }

func (w writer) write() int { return 0 }

type Pty struct{}

func (p *Pty) Close() error { return nil }

// Producer trails the type it produces, outside every gap.
func (s *Size) OpenPty() *Pty { return nil }

type infoSource interface {
	info() int
}

type staticSource struct{ n int }

// A constructor may return the interface it implements rather than the type.
func newInfoSource(n int) infoSource { return &staticSource{n: n} }

func (s *staticSource) info() int { return s.n }

type cni struct{}

func newCNI() b.Network { return &cni{} }

func (c *cni) Up() {}
