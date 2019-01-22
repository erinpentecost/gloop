package gloop_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/erinpentecost/gloop"
	"github.com/stretchr/testify/assert"
)

func TestInitialization(t *testing.T) {
	render := func(step time.Duration) error {
		return nil
	}
	simulate := func(step time.Duration) error {
		return nil
	}
	loop, err := gloop.NewLoop(render, simulate, gloop.Hz60Delay, gloop.Hz60Delay)
	assert.Nil(t, err)
	assert.NotNil(t, loop)
}

func TestInitializationError(t *testing.T) {
	render := func(step time.Duration) error {
		return nil
	}
	simulate := func(step time.Duration) error {
		return nil
	}
	loop, err := gloop.NewLoop(render, simulate, time.Duration(0), gloop.Hz60Delay)
	assert.NotNil(t, err)
	assert.Nil(t, loop)
}

func TestStartAndStop(t *testing.T) {
	render := func(step time.Duration) error {
		return nil
	}
	simulate := func(step time.Duration) error {
		return nil
	}
	loop, err := gloop.NewLoop(render, simulate, gloop.Hz60Delay, gloop.Hz60Delay)
	assert.Nil(t, err)
	assert.NotNil(t, loop)
	err = loop.Start()
	assert.Nil(t, err)
	loop.Stop(nil)
	<-loop.Done()
	assert.Nil(t, loop.Err())
}

func TestPrematureStop(t *testing.T) {
	render := func(step time.Duration) error {
		return nil
	}
	simulate := func(step time.Duration) error {
		return nil
	}
	loop, err := gloop.NewLoop(render, simulate, gloop.Hz60Delay, gloop.Hz60Delay)
	assert.Nil(t, err)
	assert.NotNil(t, loop)
	loop.Stop(nil)
	err = loop.Start()
	assert.NotNil(t, err)
	<-loop.Done()
	assert.Nil(t, loop.Err())
}

func TestDoubleStop(t *testing.T) {
	render := func(step time.Duration) error {
		return nil
	}
	simulate := func(step time.Duration) error {
		return nil
	}
	loop, err := gloop.NewLoop(render, simulate, gloop.Hz60Delay, gloop.Hz60Delay)
	assert.Nil(t, err)
	assert.NotNil(t, loop)
	err = loop.Start()
	assert.Nil(t, err)
	loop.Stop(nil)
	loop.Stop(nil)
	<-loop.Done()
	loop.Stop(nil)
	assert.Nil(t, loop.Err())
}

func TestRenderError(t *testing.T) {
	render := func(step time.Duration) error {
		return fmt.Errorf("Intentional error")
	}
	simulate := func(step time.Duration) error {
		return nil
	}
	loop, err := gloop.NewLoop(render, simulate, gloop.Hz60Delay, gloop.Hz60Delay)
	assert.Nil(t, err)
	assert.NotNil(t, loop)
	err = loop.Start()
	assert.Nil(t, err)
	<-loop.Done()
	assert.NotNil(t, loop.Err())
}

func TestSimulateError(t *testing.T) {
	render := func(step time.Duration) error {
		return nil
	}
	simulate := func(step time.Duration) error {
		return fmt.Errorf("Intentional error")
	}
	loop, err := gloop.NewLoop(render, simulate, gloop.Hz60Delay, gloop.Hz60Delay)
	assert.Nil(t, err)
	assert.NotNil(t, loop)
	err = loop.Start()
	assert.Nil(t, err)
	<-loop.Done()
	assert.NotNil(t, loop.Err())
}

func TestMetricPublication(t *testing.T) {
	render := func(step time.Duration) error {
		return nil
	}
	simulate := func(step time.Duration) error {
		return nil
	}
	loop, err := gloop.NewLoop(render, simulate, gloop.Hz60Delay, gloop.Hz60Delay)
	assert.Nil(t, err)
	assert.NotNil(t, loop)
	err = loop.Start()
	assert.Nil(t, err)

	sample := <-loop.Heartbeat()

	loop.Stop(nil)
	<-loop.Done()
	assert.Nil(t, loop.Err())

	assert.NotNil(t, sample)
}
