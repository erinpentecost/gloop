package gogameloop

import (
	"math"
	"time"
)

type statWindow struct {
	samples  []time.Duration
	curIndex int
}

func newStatWindow(samples int) statWindow {
	return statWindow{
		samples:  make([]time.Duration, samples, samples),
		curIndex: 0,
	}
}

func (p *statWindow) AddSample(sample time.Duration) {
	p.samples[p.curIndex] = sample
	p.curIndex = (p.curIndex + 1) % len(p.samples)
}

func (p *statWindow) Report() (mean, stdDev time.Duration) {
	sum := time.Duration(0)
	varNumerator := time.Duration(0)
	for _, s := range p.samples {
		sum += s
	}
	mean = sum / time.Duration(len(p.samples))
	for _, s := range p.samples {
		varNumerator += (s - mean) * (s - mean)
	}
	stdDev = time.Duration(int64(math.Sqrt(float64(varNumerator) / float64(len(p.samples)))))
	return
}

type statProfile struct {
	// arrivalWindow is how often the function is invoked.
	arrivalWindow statWindow
	// serviceWindow is how long the function takes.
	serviceWindow statWindow
	lastStart     time.Time
}

func newStatProfile(samples int) statProfile {
	now := time.Now()
	return statProfile{
		arrivalWindow: newStatWindow(samples),
		serviceWindow: newStatWindow(samples),
		lastStart:     now,
	}
}

func (p *statProfile) MarkStart() {
	now := time.Now()

	p.arrivalWindow.AddSample(now.Sub(p.lastStart))
	p.lastStart = now
}

func (p *statProfile) MarkEnd() {
	now := time.Now()

	p.serviceWindow.AddSample(now.Sub(p.lastStart))
}

func (p *statProfile) RuntimeStats() (mean, stdDev time.Duration) {
	return p.serviceWindow.Report()
}

func (p *statProfile) FrequencyStats() (mean, stdDev time.Duration) {
	return p.arrivalWindow.Report()
}

// LoopStats profiles runtime and frequency of game loop functions.
type LoopStats struct {
	LoopCount               uint64
	RenderRuntimeMean       time.Duration
	RenderRuntimeStdDev     time.Duration
	RenderFrequencyMean     time.Duration
	RenderFrequencyStdDev   time.Duration
	SimulateRuntimeMean     time.Duration
	SimulateRuntimeStdDev   time.Duration
	SimulateFrequencyMean   time.Duration
	SimulateFrequencyStdDev time.Duration
}

func newLoopStats(loopCount uint64, render *statProfile, simulate *statProfile) LoopStats {
	ls := LoopStats{}
	ls.LoopCount = loopCount
	ls.RenderFrequencyMean, ls.RenderFrequencyStdDev = render.FrequencyStats()
	ls.RenderRuntimeMean, ls.RenderRuntimeStdDev = render.RuntimeStats()
	ls.SimulateFrequencyMean, ls.SimulateFrequencyStdDev = simulate.FrequencyStats()
	ls.SimulateRuntimeMean, ls.SimulateRuntimeStdDev = simulate.RuntimeStats()
	return ls
}
