package gloop

import (
	"time"
)

const rateSampleCount int = 100

// rateTracker collects some number of samples, finds the average,
// and then publishes the average on its output channel.
type rateTracker struct {
	source         TokenSource
	lastDone       time.Time
	expectedRate   time.Duration
	samples        []time.Duration
	curIndex       int
	sampleReceiver chan PerfSample
}

func newRateTracker(source TokenSource, expectedRate time.Duration) rateTracker {
	return rateTracker{
		source:         source,
		lastDone:       time.Now(),
		expectedRate:   expectedRate,
		samples:        make([]time.Duration, 0, rateSampleCount),
		sampleReceiver: make(chan PerfSample, 1),
	}
}

func (r *rateTracker) Deadline() time.Time {
	return r.lastDone.Add(r.expectedRate)
}

func (r *rateTracker) Receive() <-chan PerfSample {
	return r.sampleReceiver
}

func (r *rateTracker) Stop() {
	close(r.sampleReceiver)
}

func (r *rateTracker) MarkDone() {
	now := time.Now()
	sample := now.Sub(r.lastDone)
	r.lastDone = now

	r.samples[r.curIndex] = sample

	r.curIndex++
	// Once we get enough samples, publish and reset.
	if r.curIndex >= cap(r.samples) {
		r.sampleReceiver <- PerfSample{
			Source:   r.source,
			Expected: r.expectedRate,
			Average:  r.Average()}
		r.samples = make([]time.Duration, 0, rateSampleCount)
		r.curIndex = 0
	}
}

func (r *rateTracker) Average() time.Duration {
	sum := time.Duration(0)
	for _, sample := range r.samples {
		sum += sample
	}
	return sum / time.Duration(len(r.samples))
}
