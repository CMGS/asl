package a

type Backend struct{}

func (b *Backend) Scan() {}

func (b *Backend) observe() {} // want `unexported method Backend\.observe declared above exported method Recover; move unexported methods below the exported set`

func (b *Backend) Converge() {}

func (b *Backend) converge() {} // want `unexported method Backend\.converge declared above exported method Recover; move unexported methods below the exported set`

func (b *Backend) Recover() {}

func (b *Backend) collect() {}

type guard struct{}

func (g guard) load() {} // want `unexported method guard\.load declared above exported method Load; move unexported methods below the exported set`

func (g guard) Load() {}

type clean struct{}

func (c clean) Get() {}

func (c clean) Put() {}

func (c clean) get() {}

func (c clean) put() {}

type mixed struct{}

func (m mixed) Alpha() {}

func (m mixed) beta() {} // want `unexported method mixed\.beta declared above exported method End; move unexported methods below the exported set`

type other struct{}

// A different receiver interleaving the partition is not itself a break.
func (o other) tail() {}

func (m mixed) End() {}

type box[T any] struct{ v T }

func (b *box[T]) peek() T { return b.v } // want `unexported method box\.peek declared above exported method Take; move unexported methods below the exported set`

func (b *box[T]) Take() T { return b.v }

func standalone() {}

func Standalone() {}

type Sandbox struct{}

func (s *Sandbox) Run() {}

func (s *Sandbox) cleanup() {}

type Pty struct{}

func (p *Pty) Close() {}

func (s *Sandbox) OpenPty() *Pty { return nil }
