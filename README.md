# gloop

[![Go Report Card](https://goreportcard.com/badge/github.com/erinpentecost/gloop)](https://goreportcard.com/report/github.com/erinpentecost/gloop)
[![Travis CI](https://travis-ci.org/erinpentecost/gloop.svg?branch=master)](https://travis-ci.org/erinpentecost/gloop.svg?branch=master)
[![GoDoc](https://godoc.org/github.com/erinpentecost/gloop?status.svg)](https://godoc.org/github.com/erinpentecost/gloop)

Real-ish time Go simulation loop with support for simultaneous fixed step (simulation) and elastic step (rendering) functions. Includes a heartbeat channel to monitor loop health and performance metrics. 

## Example

A full example using Vulkan and metrics publishing is availabe in [gloopex](https://github.com/erinpentecost/gloopex) repo.

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
// At some point, call Stop.
// It's safe to do this from inside render() or simulate().
loop.Stop(nil)
// Wait for the loop to finish.
// Once this chan is closed, it is guaranteed that the
// neither render() nor simulate() will be called again.
<-loop.Done()
```

## Install

```go
go get -u github.com/erinpentecost/gloop
```

## Testing

```sh
# You can use normal go tooling...
go test ./...
# Or the make file...
make test
```
