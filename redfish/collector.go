package redfish

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "hw"

// Client is an interface of a client for Redfish API.
type Client interface {
	traverse(ctx context.Context) dataMap
}

type dataMap map[string]*gabs.Container

// Collector implements prometheus.Collector interface.
type Collector struct {
	rule    *CollectRule
	client  Client
	dataMap atomic.Value
}

// NewCollector returns a new instance of Collector.
func NewCollector(rule *CollectRule, client Client) (*Collector, error) {
	return &Collector{rule: rule, client: client}, nil
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
	ctx1, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	dataMap := c.client.traverse(ctx1)
	c.dataMap.Store(dataMap)
}
