package collector

import (
	"sync"

	"github.com/cybozu-go/setup-hw/redfish"
	"github.com/prometheus/client_golang/prometheus"
)

// Namespace is the first part of the metrics name.
const Namespace = "hw"

// Collector is the interface for collectors.
type Collector interface {
	// Collect exposes new metrics via prometheus registry.
	Collect(ch chan<- prometheus.Metric) error
}

// SafeMetrics is cached metrics.
type SafeMetrics struct {
	metrics map[string][]redfish.Value
	mux     sync.Mutex
}

// Metrics is a instance of cached metrics.
var Metrics *SafeMetrics

// NewSafeMetrics return a new instance of SafeMetrics.
func NewSafeMetrics() *SafeMetrics {
	return &SafeMetrics{metrics: make(map[string][]redfish.Value)}
}

// Set updates cached metrics.
func (m *SafeMetrics) Set(name string, values []redfish.Value) {
	m.mux.Lock()
	m.metrics[name] = values
	m.mux.Unlock()
}

// Get returns cached metrics.
func (m *SafeMetrics) Get(name string) []redfish.Value {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.metrics[name]
}
