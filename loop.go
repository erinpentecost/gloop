// Package gloop implements a game loop.
package gloop

import (
	"sync"
	"time"
)

// Hz60Delay is 1/60th of a second.
const Hz60Delay time.Duration = time.Duration(int64(time.Second) / 60)

type state int

const (
	stateInit state = iota
	stateRun  state = iota
	stateStop state = iota
)

// LoopFn is a function that is called inside the game loop.
// step should be treated as if it was the amount of time that
// elapsed since the last call.
type LoopFn func(step time.Duration) error

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
	mu             sync.Mutex
	done           chan interface{}
	err            error
	heartbeat      chan LatencySample
	curState       state
}

// NewLoop creates a new game loop.
func NewLoop(Render, Simulate LoopFn, RenderRate, SimulationRate time.Duration) (*Loop, error) {
	// Input validation.
	if RenderRate <= 0 {
		return nil, wrapLoopError(nil, TokenLoop, "RenderRate can't be lte 0")
	}
	if SimulationRate <= 0 {
		return nil, wrapLoopError(nil, TokenLoop, "SimulationRate can't be lte 0")
	}

	// Init loop.
	return &Loop{
		Render:         Render,
		Simulate:       Simulate,
		SimulationRate: SimulationRate,
		RenderRate:     RenderRate,
		done:           make(chan interface{}),
		err:            nil,
		heartbeat:      make(chan LatencySample),
		curState:       stateInit,
	}, nil
}

// Heartbeat returns the heartbeat channel which
// can be used to monitor the health of the game loop.
// A pulse will be sent every second with current simulation
// and render latency.
func (l *Loop) Heartbeat() <-chan LatencySample {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.heartbeat
}

// Done returns a chan that indicates when the loop is stopped.
// When this finishes, you should do cleanup.
func (l *Loop) Done() <-chan interface{} {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.done
}

// Stop halts the loop and sets Err().
// You probably want to make a call to this somewhere in Simulate().
func (l *Loop) Stop(err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.curState != stateStop {
		close(l.done)
		l.err = err
		l.curState = stateStop
	}
}

// Err returns the the reason why the loop closed if there was an error.
// Err will return nil if the loop has not yet run, is currently running,
// or closed without an error.
func (l *Loop) Err() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.err
}

// Start initiates a game loop. This call does not block.
// To stop the loop, close the done chan.
// To get notified before Simulate or Render are called, pull items from
// the heartbeat channel.
// If either Render or Simulate throw an error, the error will be made available
// on the output error channel and the loop will stop.
func (l *Loop) Start() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Silently fail on re-starts.
	if l.curState != stateInit {
		return wrapLoopError(nil, TokenLoop, "Loop is already running or is done")
	}
	l.curState = stateRun

	// Time tracking.
	previousTime := time.Now()
	simAccumulator := time.Duration(0)
	rendAccumulator := time.Duration(0)

	// Stats heartbeat channel set up
	heartTick := time.NewTicker(time.Second)
	sendBeat := func(ps LatencySample) {
		select {
		case l.heartbeat <- ps:
		default: // Throw it away if no one is listening.
		}
	}

	// tick goes off often enough that both l.SimulationRate and l.RenderRate will be invoked
	// when they expect to, and no earlier.
	tick := time.NewTicker(gcd(l.SimulationRate, l.RenderRate))

	go func() {
		defer tick.Stop()
		defer heartTick.Stop()
		defer close(l.heartbeat)
		defer l.Stop(nil)

		simLatency := newLatencyTracker()
		rendLatency := newLatencyTracker()

		for {
			select {
			case <-l.Done():
				break
			case <-heartTick.C:
				sendBeat(LatencySample{
					RenderLatency:   rendLatency.Latency(),
					SimulateLatency: simLatency.Latency(),
				})
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

					// Actually call simulate...
					if er := l.Simulate(l.SimulationRate); er != nil {
						wrapped := wrapLoopError(er, TokenSimulate, "Error returned by Simulate(%s)", l.SimulationRate.String())
						wrapped.Misc["curTime"] = curTime
						l.Stop(wrapped)
						break
					}

					simLatency.MarkDone(l.SimulationRate)

					// Keep track of leftover time.
					simAccumulator -= l.SimulationRate
				}

				// Run the render function. Only do it once, though.
				// This lets me have an upper limit on FPS.
				if rendAccumulator >= l.RenderRate {

					// Actually call render...
					leftOver := rendAccumulator % l.RenderRate
					work := rendAccumulator - leftOver
					if er := l.Render(work); er != nil {
						wrapped := wrapLoopError(er, TokenRender, "Error returned by Render(%s)", time.Duration(work).String())
						wrapped.Misc["curTime"] = curTime
						l.Stop(wrapped)
						break
					}

					rendLatency.MarkDone(work)

					// Keep track of leftover time.
					rendAccumulator = leftOver
				}
			}
		}
	}()
	return nil
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
