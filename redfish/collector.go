package redfish

import (
	"context"
	"os"
	"sync"

	"github.com/Jeffail/gabs"
	"github.com/prometheus/client_golang/prometheus"
	yaml "gopkg.in/yaml.v2"
)

// Namespace is the first part of the metrics name.
const Namespace = "hw"

type RedfishCollector struct {
	rules []ConvertRule
}

func (c *RedfishCollector) Init(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	var rules []ConvertRule
	err = yaml.NewDecoder(f).Decode(&rules)
	if err != nil {
		return err
	}
	c.rules = rules
	return nil
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
