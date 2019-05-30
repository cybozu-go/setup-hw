package redfish

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/cybozu-go/log"
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
	rule                          *CollectRule
	client                        Client
	dataMap                       atomic.Value
	lastUpdate                    time.Time
	lastUpdateDesc                *prometheus.Desc
	lastUpdateDurationMinutesDesc *prometheus.Desc
}

// NewCollector returns a new instance of Collector.
func NewCollector(rule *CollectRule, client Client) (*Collector, error) {
	desc1 := prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "last_update"), "", nil, nil)
	desc2 := prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "last_update_duration_minutes"), "", nil, nil)
	return &Collector{rule: rule, client: client, lastUpdateDesc: desc1, lastUpdateDurationMinutesDesc: desc2}, nil
}

// Describe sends descriptions of metrics.
func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, rule := range c.rule.MetricRules {
		for _, propertyRule := range rule.PropertyRules {
			ch <- propertyRule.desc
		}
	}
	ch <- c.lastUpdateDesc
	ch <- c.lastUpdateDurationMinutesDesc
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
	m, err := prometheus.NewConstMetric(c.lastUpdateDesc, prometheus.CounterValue, float64(c.lastUpdate.Unix()))
	if err != nil {
		log.Error("failed to register last_update", map[string]interface{}{log.FnError: err})
	}
	ch <- m
	m, err = prometheus.NewConstMetric(c.lastUpdateDurationMinutesDesc, prometheus.GaugeValue, time.Now().Sub(c.lastUpdate).Minutes())
	if err != nil {
		log.Error("failed to register last_update_duration_minutes", map[string]interface{}{log.FnError: err})
	}
	ch <- m
}

// Update collects metrics from BMCs via Redfish.
func (c *Collector) Update(ctx context.Context) {
	ctx1, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	dataMap := c.client.traverse(ctx1)
	c.dataMap.Store(dataMap)
	c.lastUpdate = time.Now()
}
