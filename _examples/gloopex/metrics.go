package main

import (
	"context"
	"expvar"
	"time"

	"github.com/erinpentecost/gloop"

	"net/http"

	"github.com/zserge/metric"
)

// MetricsServer collects and publishes metrics.
type MetricsServer struct {
	renderLatency   metric.Metric
	simulateLatency metric.Metric
}

// NewMetricsServer creates a new metrics server.
func NewMetricsServer() MetricsServer {
	return MetricsServer{
		renderLatency:   metric.NewGauge("5m5s"),
		simulateLatency: metric.NewGauge("5m5s"),
	}
}

// Serve starts an http server.
func (m *MetricsServer) Serve(done <-chan interface{}) {
	expvar.Publish("RenderLatencyMs", m.renderLatency)
	expvar.Publish("SimulateLatencyMs", m.simulateLatency)

	server := &http.Server{Addr: ":8000", Handler: metric.Handler(metric.Exposed)}

	// Start hosting http nonblocking
	go func() {
		server.ListenAndServe()
	}()

	// Wait for cancellation and then shutdown http
	go func() {
		<-done
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()
}

// Publish takes in some sample.
func (m *MetricsServer) Publish(sample gloop.LatencySample) {
	m.renderLatency.Add(float64(sample.RenderLatency) / 1000000.0)
	m.simulateLatency.Add(float64(sample.SimulateLatency) / 1000000.0)
}
