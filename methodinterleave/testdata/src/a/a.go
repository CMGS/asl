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
