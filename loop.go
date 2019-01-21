// Package gloop implements a game loop.
package gloop

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
	// earlier. It will run 0 or 1 times per game loop on an
	// elastic step.
	Render LoopFn
	// Simulate will attempt be called at SimulationRate with
	// a fixed step. It may be executed more often than Render per
	// game loop.
	Simulate LoopFn
	// RenderRate controls how often Render will be called.
	// This is the time delay between calls.
	RenderRate time.Duration
	// SimulationRate controls how often Simulate will be called.
	// This is the time delay between calls.
	SimulationRate time.Duration
}

// NewLoop creates a new game loop.
func NewLoop(Render, Simulate LoopFn, RenderRate, SimulationRate time.Duration) Loop {
	return Loop{
		Render:         Render,
		Simulate:       Simulate,
		SimulationRate: SimulationRate,
		RenderRate:     RenderRate,
	}
}

// Start initiates a game loop. This call does not block.
// To stop the loop, close(ctx.Done).
// To get notified before Simulate or Render are called, pull items from
// the LoopMetric channel.
// If either Render or Simulate throw an error, the error will be made available
// on the output error channel and the loop will stop.
func (l *Loop) Start(ctx context.Context) (<-chan LoopMetric, <-chan error) {
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
	statHeartbeat := make(chan LoopMetric, 1)
	loopCount := uint64(0)
	sendPulse := func(sample LoopMetric) {
		select {
		case statHeartbeat <- sample:
		default:
		}
	}

	// Input validation.
	if l.RenderRate <= 0 {
		defer close(errc)
		defer close(statHeartbeat)
		sendError(wrapLoopError(nil, TokenLoop, "RenderRate can't be lte 0."))
		return statHeartbeat, errc
	}
	if l.SimulationRate <= 0 {
		defer close(errc)
		defer close(statHeartbeat)
		sendError(wrapLoopError(nil, TokenLoop, "SimulationRate can't be lte 0."))
		return statHeartbeat, errc
	}

	// tick goes off often enough that both l.SimulationRate and l.RenderRate will be invoked
	// when they expect to, and no earlier.
	tick := time.NewTicker(gcd(l.SimulationRate, l.RenderRate))

	go func() {
		defer tick.Stop()
		defer close(statHeartbeat)
		defer close(errc)

		lastSimulate := time.Now()
		lastRender := time.Now()

		for {
			select {
			case <-ctx.Done():
				break
			case <-tick.C:
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

					// Publish metric...
					simNow := time.Now()
					simRate := simNow.Sub(lastSimulate)
					lastSimulate = simNow
					sendPulse(LoopMetric{TokenSimulate, simRate, loopCount})

					// Actually call simulate...
					if er := l.Simulate(ctx, l.SimulationRate); er != nil {
						wrapped := wrapLoopError(er, TokenSimulate, "Error returned by Simulate(ctx, %s).", l.SimulationRate.String())
						wrapped.Misc["loopCount"] = loopCount
						wrapped.Misc["curTime"] = curTime
						wrapped.Misc["ctx"] = ctx
						sendError(wrapped)
						break
					}

					// Keep track of leftover time.
					simAccumulator -= l.SimulationRate
				}

				// Run the render function. Only do it once, though.
				// This lets me have an upper limit on FPS.
				if rendAccumulator >= l.RenderRate {
					// Publish metric...
					rendNow := time.Now()
					rendRate := rendNow.Sub(lastRender)
					lastRender = rendNow
					sendPulse(LoopMetric{TokenSimulate, rendRate, loopCount})

					// Actually call render...
					leftOver := rendAccumulator % l.RenderRate
					if er := l.Render(ctx, rendAccumulator-leftOver); er != nil {
						wrapped := wrapLoopError(er, TokenRender, "Error returned by Render(ctx, %s).", time.Duration(rendAccumulator-leftOver).String())
						wrapped.Misc["loopCount"] = loopCount
						wrapped.Misc["curTime"] = curTime
						wrapped.Misc["ctx"] = ctx
						sendError(wrapped)
						break
					}

					// Keep track of leftover time.
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
