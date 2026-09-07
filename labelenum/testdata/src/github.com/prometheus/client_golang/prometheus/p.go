package prometheus

type CounterVec struct{}

func NewCounterVec(labels []string) *CounterVec { return &CounterVec{} }
