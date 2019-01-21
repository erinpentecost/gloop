package gogameloop

import "time"

// LoopMetric is some gauge for actual rate.
type LoopMetric struct {
	Source   TokenSource
	Duration time.Duration
	Frame    uint64
}
