package b

import "time"

func signed(d time.Duration) time.Duration {
	if d > 0 { // want `if/return fallback on d is cmp.Or\(d, time.Second\)`
		return d
	}
	return time.Second
}
