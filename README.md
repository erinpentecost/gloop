<p align="center"><img src="icon.jpg" alt="gloop" width="600"/></p>

[![Go Report Card](https://goreportcard.com/badge/github.com/erinpentecost/gloop)](https://goreportcard.com/report/github.com/erinpentecost/gloop)
[![Travis CI](https://travis-ci.org/erinpentecost/gloop.svg?branch=master)](https://travis-ci.org/erinpentecost/gloop.svg?branch=master)
[![GoDoc](https://godoc.org/github.com/erinpentecost/gloop?status.svg)](https://godoc.org/github.com/erinpentecost/gloop)

Real-ish time Go simulation loop with support for simultaneous fixed step (simulation) and elastic step (rendering) functions. Includes a heartbeat channel to monitor loop health and performance metrics. 

## Example

A full example using Vulkan and metrics publishing is availabe in [_examples/gloopex](_examples/gloopex/README.md) folder.

```go
render := func(step time.Duration) error {
    // Do elastic-step work here.
    // The step param will be larger than RenderLatency
    // if we start to fall behind. 
    return nil
}
simulate := func(step time.Duration) error {
    // Do fixed-step work here.
    // The step param will always be SimulationLatency,
    // and we'll invoke the function more often if
    // we start to fall behind.
    return nil
}
// Set up the loop here.
loop, err := gloop.NewLoop(render, simulate, gloop.Hz60Delay, gloop.Hz60Delay)
// Start up the loop. This is not blocking.
loop.Start()
// Wait some period of time...
<-time.NewTimer(time.Minute).C
// At some point, call Stop.
// It's safe to do this from inside render() or simulate().
// The loop will also stop if render() or simulate() return an error.
loop.Stop(nil)
// Wait for the loop to finish.
// Once this chan is closed, it is guaranteed that
// neither render() nor simulate() will be called again.
<-loop.Done()
```
## Quick Tutorial

`loop.Start(...)` starts the loop in a different goroutine.

`loop.Stop(...)` will halt the loop. This is thread safe, and can be called from within `loop.Render(...)` or `loop.Simulate(...)`.

If `loop.Render(...)` or `loop.Simulate(...)` return an error, the loop will halt and the loop's `loop.Err()` will be set to non-nil.

Pull performance metrics out of the loop with `sample <- loop.Heartbeat()`.

Wait for the loop to finish with `<- loop.Done()`. Once this closes, `loop.Render(...)` and `loop.Simulate(...)` will not be called again. The chan will also not close until any currently-executing calls to either of those functions finish.

There is no need to use synchronization objects; only one call to `loop.Render(...)` or `loop.Simulate(...)` will run at a time.

## Install

```go
go get -u github.com/erinpentecost/gloop
```

## Testing

```sh
# You can use normal go tooling...
go test ./...
# Or the makefile...
make test
```
