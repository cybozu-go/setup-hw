package redfish

import (
	"context"
	"sync/atomic"

	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "hw"

type dataMap map[string]*gabs.Container

// Collector implements prometheus.Collector interface.
type Collector struct {
	rule    *CollectRule
	client  *client
	dataMap atomic.Value
}

// CollectorConfig is a set of configurations for Collector.
type CollectorConfig struct {
	AddressConfig *config.AddressConfig
	Port          string
	UserConfig    *config.UserConfig
	Rule          *CollectRule
}

// NewCollector returns a new instance of Collector.
func NewCollector(cc *CollectorConfig) (*Collector, error) {
	client, err := newClient(cc)
	if err != nil {
		return nil, err
	}

	return &Collector{rule: cc.Rule, client: client}, nil
}

// Describe sends descriptions of metrics.
func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, rule := range c.rule.MetricRules {
		for _, propertyRule := range rule.PropertyRules {
			ch <- propertyRule.desc
		}
	}
}

// Collect sends metrics collected from BMC via Redfish.
func (c Collector) Collect(ch chan<- prometheus.Metric) {
	v := c.dataMap.Load()
	if v == nil {
		return
	}
	dataMap := v.(dataMap)

	for _, rule := range c.rule.MetricRules {
		metrics := rule.matchDataMap(dataMap)
		for _, m := range metrics {
			ch <- m
		}
	}
}

// Update collects metrics from BMCs via Redfish.
func (c *Collector) Update(ctx context.Context) {
	dataMap := c.client.traverse(ctx)
	c.dataMap.Store(dataMap)
}
