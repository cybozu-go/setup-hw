package redfish

import (
	"context"
	"sync"

	"github.com/Jeffail/gabs"
	"github.com/cybozu-go/setup-hw/collector"
	"github.com/prometheus/client_golang/prometheus"
)

// Namespace is the first part of the metrics name.
const Namespace = "hw"

type RedfishCollector struct {
	collectors map[string]collector.Collector
}

func (c RedfishCollector) Describe(ch chan<- *prometheus.Desc) {
}

func (c RedfishCollector) Collect(ch chan<- prometheus.Metric) {
	// TODO
	// values := Metrics.Get("chassis")
	// for _, value := range values {
	// 	float, err := strconv.ParseFloat(value.Value, 64)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	c.current = prometheus.NewDesc(
	// 		prometheus.BuildFQName(Namespace, "", value.Name),
	// 		"Overall status of chassis components.",
	// 		nil, value.Labels)
	// 	ch <- prometheus.MustNewConstMetric(
	// 		c.current, prometheus.GaugeValue, float)
	// }
	// return nil
}

type ContainerMap map[string]*gabs.Container

type CachedContainer struct {
	containerMap ContainerMap
	mux          sync.Mutex
}

var Containers *CachedContainer

func init() {
	Containers = &CachedContainer{containerMap: make(ContainerMap)}
}

func (c *CachedContainer) Set(cm ContainerMap) {
	c.mux.Lock()
	c.containerMap = cm
	c.mux.Unlock()
}

func (c *CachedContainer) Get() ContainerMap {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.containerMap
}

func Update(ctx context.Context, rootPath str) error {
	data, err := Traverse(ctx, rootPath)
	if err != nil {
		return err
	}

}
