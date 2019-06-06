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
	traverse(ctx context.Context, rule *CollectRule) collected
}

type collected struct {
	data map[string]*gabs.Container
	rule *CollectRule
}

// Collector implements prometheus.Collector interface.
type Collector struct {
	ruleGetter                    RuleGetter
	client                        Client
	collected                     atomic.Value
	lastUpdate                    time.Time
	lastUpdateDesc                *prometheus.Desc
	lastUpdateDurationMinutesDesc *prometheus.Desc
}

// NewCollector returns a new instance of Collector.
func NewCollector(ruleGetter RuleGetter, client Client) (*Collector, error) {
	desc1 := prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "last_update"), "", nil, nil)
	desc2 := prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "last_update_duration_minutes"), "", nil, nil)
	return &Collector{ruleGetter: ruleGetter, client: client, lastUpdateDesc: desc1, lastUpdateDurationMinutesDesc: desc2}, nil
}

// Describe sends descriptions of metrics.
func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.lastUpdateDesc
	ch <- c.lastUpdateDurationMinutesDesc
}

// Collect sends metrics collected from BMC via Redfish.
func (c Collector) Collect(ch chan<- prometheus.Metric) {
	m := prometheus.MustNewConstMetric(c.lastUpdateDesc, prometheus.CounterValue, float64(c.lastUpdate.Unix()))
	ch <- m
	m = prometheus.MustNewConstMetric(c.lastUpdateDurationMinutesDesc, prometheus.GaugeValue, time.Now().Sub(c.lastUpdate).Minutes())
	ch <- m

	v := c.collected.Load()
	if v == nil {
		return
	}
	cl := v.(collected)

	for _, rule := range cl.rule.MetricRules {
		metrics := rule.matchDataMap(cl)
		for _, m := range metrics {
			ch <- m
		}
	}
}

// Update collects metrics from BMCs via Redfish.
func (c *Collector) Update(ctx context.Context) {
	log.Info("getting rule", nil)
	rule, err := c.ruleGetter()
	if err != nil {
		log.Error("failed to get rule", map[string]interface{}{
			log.FnError: err,
		})
		return
	}
	log.Info("start update", nil)
	ctx1, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	cl := c.client.traverse(ctx1, rule)
	c.collected.Store(cl)
	c.lastUpdate = time.Now()
	log.Info("finish update", nil)
}
