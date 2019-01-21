package gloop

import (
	"time"
)

type latencyTracker struct {
	start        time.Time
	finishedWork time.Duration
}

func newLatencyTracker() latencyTracker {
	return latencyTracker{
		start:        time.Now(),
		finishedWork: time.Duration(0),
	}
}

func (lt *latencyTracker) MarkDone(workDone time.Duration) {
	lt.finishedWork += workDone
}

func (lt *latencyTracker) Latency() time.Duration {
	// Latency is the difference between now and how far we got earlier.
	now := time.Now()
	current := lt.start.Add(lt.finishedWork)
	latency := now.Sub(current)
	// Shift the start period and current finishedWork so I don't
	// end up dealing with massive numbers. Probably not necessary,
	// but this will prevent rollover.
	lt.start = now.Add(-1 * latency)
	lt.finishedWork = time.Duration(0)
	return latency
}
