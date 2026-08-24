package a

type Provider struct{}

func (p *Provider) Start() {}

func appendArg(args []string) []string { return args } // want `standalone function appendArg declared between Provider methods; keep the method set contiguous and move it above or below`

func (p *Provider) Exec() {}

type record struct{ pid int } // want `type record declared between Provider methods; keep the method set contiguous and move it above or below`

func (p *Provider) Stop() {}

func (p *Provider) NewHelper() *Provider { return p }

func apply(r *record) int { return r.pid } // want `standalone function apply declared between Provider methods; keep the method set contiguous and move it above or below`

func (p *Provider) Close() {}

type Queue struct{}

func (q *Queue) Push() {}

type (
	kind  string // want `type kind declared between Queue methods; keep the method set contiguous and move it above or below`
	label string
)

func (q *Queue) Pop() {}

type Engine struct{}

func (e *Engine) Configure() {}

type lease struct{}

func (l *lease) close() {}

func (e *Engine) Launch() {} // want `method Engine.Launch resumes the Engine method set after type lease; keep the method set contiguous, only producers trail a foreign type block`

func (e *Engine) NewLease() *lease { return nil }
