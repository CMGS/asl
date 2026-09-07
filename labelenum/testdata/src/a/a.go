package a

import "github.com/prometheus/client_golang/prometheus"

var (
	Ops = prometheus.NewCounterVec(
		// result=ok|failed|skipped; reason=adopted|noop // want `comment enumerates the values of result`
		[]string{"op", "result", "reason"},
	)

	Boot = prometheus.NewCounterVec([]string{"mode"}) // mode=run|clone // want `comment enumerates the values of mode`

	// Wake counts wake outcomes by result.
	Wake = prometheus.NewCounterVec([]string{"result"})
)
