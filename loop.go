// Package gogameloop implements a game loop.
package gogameloop

import (
	"context"
	"time"
)

// Hz60Delay is 1/60th of a second.
const Hz60Delay time.Duration = time.Duration(int64(time.Second) / 60)

// LoopFn is a function that is called inside the game loop.
// step should be treated as if it was the amount of time that
// elapsed since the last call.
type LoopFn func(ctx context.Context, step time.Duration) error

// Loop is a game loop.
type Loop struct {
	// Render will attempt be called at RenderRate, and no
	// earlier. It will run 0 or 1 times per game loop.
	Render LoopFn
	// Simulate will attempt be called at SimulationRate.
	// It will be invoked many times at a fixed step
	// if we are falling behind. It will run 0 or more
	// times per game loop.
	Simulate LoopFn
	// RenderRate controls how often Render will be called.
	// This is the time delay between calls.
	RenderRate time.Duration
	// SimulationRate controls how often Simulate will be called.
	// This is the time delay between calls.
	SimulationRate time.Duration
	// ReportRate controls how often profiling stats will be published on the heartbeat channel.
	// This is the time delay between calls.
	ReportRate time.Duration
}

// NewLoop creates a new game loop.
func NewLoop(Render, Simulate LoopFn, RenderRate, SimulationRate, ReportRate time.Duration) Loop {
	return Loop{
		Render:         Render,
		Simulate:       Simulate,
		SimulationRate: SimulationRate,
		RenderRate:     RenderRate,
		ReportRate:     ReportRate,
	}
}

// Start initiates a game loop. This call does not block.
// To stop the loop, close(ctx.Done).
// To get periodic stats on the loop, pull from the first returned channel.
// If either Render or Simulate throw an error, the error will be made available
// on the output error channel and the goroutine will stop.
func (l *Loop) Start(ctx context.Context) (<-chan LoopStats, <-chan error) {
	// Error capture.
	errc := make(chan error, 1)
	sendError := func(er error) {
		select {
		case errc <- er:
		default:
		}
	}

	// Time tracking.
	previousTime := time.Now()
	simAccumulator := time.Duration(0)
	rendAccumulator := time.Duration(0)

	// Stats heartbeat channel set up
	statHeartbeat := make(chan LoopStats)
	simStats := newStatProfile(10)
	rendStats := newStatProfile(10)
	loopCount := uint64(0)
	sendPulse := func() {
		select {
		case statHeartbeat <- newLoopStats(loopCount, &rendStats, &simStats):
		default:
		}
	}
	heartBeatTick := time.Tick(l.ReportRate)

	// tick goes off often enough that both l.SimulationRate and l.RenderRate will be invoked
	// when they expect to, and no earlier.
	tick := time.Tick(gcd(l.SimulationRate, l.RenderRate))

	go func() {
		defer close(statHeartbeat)
		defer close(errc)
		for {
			select {
			case <-ctx.Done():
				break
			case <-heartBeatTick:
				sendPulse()
			case <-tick:
				// Find delta since last frame
				curTime := time.Now()
				frameTime := curTime.Sub(previousTime)
				previousTime = curTime
				simAccumulator += frameTime
				rendAccumulator += frameTime

				// Handle simulation function.
				// This may be invoked many times.
				for simAccumulator >= l.SimulationRate {
					// Run the simulation with a fixed step.
					simStats.MarkStart()
					if er := l.Simulate(ctx, l.SimulationRate); er != nil {
						sendError(er)
						break
					}
					simStats.MarkEnd()
					simAccumulator -= l.SimulationRate
				}

				// Run the render function. Only do it once, though.
				// This lets me have an upper limit on FPS.
				if rendAccumulator >= l.RenderRate {
					leftOver := rendAccumulator % l.RenderRate
					rendStats.MarkStart()
					if er := l.Render(ctx, rendAccumulator-leftOver); er != nil {
						sendError(er)
						break
					}
					rendStats.MarkEnd()
					rendAccumulator = leftOver
				}
			}
			loopCount++
		}
	}()

	return statHeartbeat, errc
}

// gcd finds the greatest common denominator between a and b.
func gcd(a, b time.Duration) time.Duration {
	for a != b {
		if a > b {
			a -= b
		} else {
			b -= a
		}
	}
	return a
}
