package gloop

import (
	"time"
)

// PerfSample holds actual rates.
type PerfSample struct {
	Source   TokenSource
	Expected time.Duration
	Average  time.Duration
}
