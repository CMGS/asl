package a

type closer interface{ Close() error }

var _ closer = (*handle)(nil)

type handle struct{}

func (h *handle) Close() error { return nil }

var (
	_ closer = (*multi)(nil)
	_ any    = multi{}
)

type multi struct{}

func (m multi) Close() error { return nil }

var _ closer = (*far)(nil) // want `interface check for far away from its type; place it immediately above the type`

type between struct{}

type far struct{}

func (f *far) Close() error { return nil }

type below struct{}

func (b *below) Close() error { return nil }

var _ closer = (*below)(nil) // want `interface check for below away from its type; place it immediately above the type`

var _ closer = (*handle)(nil) // want `interface check for handle away from its type; place it immediately above the type`
