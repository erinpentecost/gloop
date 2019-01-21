package gloop

import (
	"time"
)

// LatencySample is a measure of how far behind simulate() or render() are.
type LatencySample struct {
	RenderLatency   time.Duration
	SimulateLatency time.Duration
}
