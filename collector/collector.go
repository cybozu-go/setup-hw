package collector

import (
	"context"
	"fmt"
	"strings"
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

	// Update gathers new metrics.
	Update(ctx context.Context) error
}

// Factories contains the list of all available collectors.
var Factories = make(map[string]func() Collector)

// SafeMetrics is cached metrics.
type SafeMetrics struct {
	metrics map[string][]redfish.Value
	mux     sync.Mutex
}

// Metrics is a instance of cached metrics.
var Metrics *SafeMetrics

func init() {
	Metrics = newSafeMetrics()
}

// newSafeMetrics return a new instance of SafeMetrics.
func newSafeMetrics() *SafeMetrics {
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

// LoadCollectors reads comma-separated collector names
// and returns collectors indexed by their names.
func LoadCollectors(list string) (map[string]Collector, error) {
	collectors := map[string]Collector{}
	for _, name := range strings.Split(list, ",") {
		fn, ok := Factories[name]
		if !ok {
			return nil, fmt.Errorf("collector '%s' not available", name)
		}
		collectors[name] = fn()
	}
	return collectors, nil
}
