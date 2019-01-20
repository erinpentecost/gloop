// Package gogameloop implements a game loop.
package gogameloop

import (
	"context"
	"time"
)

// LoopFn is a function that is called inside the game loop.
// step should be treated as if it was the amount of time that
// elapsed since the last call.
type LoopFn func(ctx context.Context, step time.Duration)

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
	// SimulationRate controls how often Simulate will be called.
	SimulationRate time.Duration
	// RenderRate controls how often Render will be called.
	RenderRate time.Duration
}

// Start initiates a game loop. This call does not block.
// To stop the loop, close(ctx.Done).
// The returned channel will be pulsed whenever a Simulate() or Render() is going to be called.
// The content of that channel includes profiling statistics on those functions.
func (l *Loop) Start(ctx context.Context) <-chan LoopStats {
	// Stats heartbeat channel set up
	statHeartbeat := make(chan LoopStats, 1)
	defer close(statHeartbeat)
	simStats := newStatProfile(10)
	rendStats := newStatProfile(10)
	loopCount := uint64(0)
	sendPulse := func() {
		select {
		case statHeartbeat <- newLoopStats(loopCount, &rendStats, &simStats):
		default:
		}
	}

	// Now keep track of timing so I know when to invoke simulate or render.
	previousTime := time.Now()
	simAccumulator := time.Duration(0)
	rendAccumulator := time.Duration(0)

	// tick goes off often enough that both l.SimulationRate and l.RenderRate will be invoked
	// when they expect to, and no earlier.
	tick := time.Tick(gcd(l.SimulationRate, l.RenderRate))

	go func() {
		for {
			select {
			case <-ctx.Done():
				break
			case <-tick:
				// Find delta since last frame
				curTime := time.Now()
				frameTime := curTime.Sub(previousTime)
				previousTime = curTime
				simAccumulator += frameTime
				rendAccumulator += frameTime

				// If I'm going to do some work, first pulse the heartbeat.
				if (simAccumulator >= l.SimulationRate) || (rendAccumulator >= l.RenderRate) {
					sendPulse()
				}

				// Handle simulation function.
				// This may be invoked many times.
				for simAccumulator >= l.SimulationRate {
					// Run the simulation with a fixed step.
					simStats.MarkStart()
					l.Simulate(ctx, l.SimulationRate)
					simStats.MarkEnd()
					simAccumulator -= l.SimulationRate
				}

				// Run the render function. Only do it once, though.
				// This lets me have an upper limit on FPS.
				if rendAccumulator >= l.RenderRate {
					leftOver := rendAccumulator % l.RenderRate
					rendStats.MarkStart()
					l.Render(ctx, rendAccumulator-leftOver)
					rendStats.MarkEnd()
					rendAccumulator = leftOver
				}

				// Report stats.
			}
			loopCount++
		}
	}()

	return statHeartbeat
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
