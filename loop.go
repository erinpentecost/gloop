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

// Start initiates a game loop.
func (l *Loop) Start(ctx context.Context) {
	previousTime := time.Now()
	simAccumulator := time.Nanosecond * 0
	rendAccumulator := time.Nanosecond * 0

	// tick goes off often enough that both l.SimulationRate and l.RenderRate will be invoked
	// when the expect to, and no earlier.
	tick := time.Tick(time.Duration(gcd(l.SimulationRate.Nanoseconds(), l.RenderRate.Nanoseconds())))

	for {
		select {
		case <-ctx.Done():
			return
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
				l.Simulate(ctx, l.SimulationRate)
				simAccumulator -= l.SimulationRate
			}

			// Run the render function. Only do it once, though.
			// This lets me have an upper limit on FPS.
			if rendAccumulator >= l.RenderRate {
				leftOver := rendAccumulator % l.RenderRate
				l.Render(ctx, rendAccumulator-leftOver)
				rendAccumulator = leftOver
			}
		}
	}
}

// gcd finds the greatest common denominator between a and b.
func gcd(a, b int64) int64 {
	for a != b {
		if a > b {
			a -= b
		} else {
			b -= a
		}
	}
	return a
}
