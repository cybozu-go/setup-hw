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
	Traverse(ctx context.Context, rule *CollectRule) Collected
	GetVersion(ctx context.Context) (string, error)
}

// NewCollected creates a new Collected instance
func NewCollected(data map[string]*gabs.Container, rule *CollectRule) *Collected {
	return &Collected{
		data: data,
		rule: rule,
	}
}

// Collected represents the collected data from Redfish
type Collected struct {
	data map[string]*gabs.Container
	rule *CollectRule
}

// Data returns the collected data
func (c Collected) Data() map[string]*gabs.Container {
	return c.data
}

// Rule returns the rul.
func (c Collected) Rule() *CollectRule {
	return c.rule
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

// Collect sends metrics Collected from BMC via Redfish.
func (c Collector) Collect(ch chan<- prometheus.Metric) {
	m := prometheus.MustNewConstMetric(c.lastUpdateDesc, prometheus.CounterValue, float64(c.lastUpdate.Unix()))
	ch <- m
	m = prometheus.MustNewConstMetric(c.lastUpdateDurationMinutesDesc, prometheus.GaugeValue, time.Now().Sub(c.lastUpdate).Minutes())
	ch <- m

	v := c.collected.Load()
	if v == nil {
		return
	}
	cl := v.(Collected)

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
	rule, err := c.ruleGetter(ctx)
	if err != nil {
		log.Error("failed to get rule", map[string]interface{}{
			log.FnError: err,
		})
		return
	}
	log.Info("start update", nil)
	ctx1, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	cl := c.client.Traverse(ctx1, rule)
	c.collected.Store(cl)
	c.lastUpdate = time.Now()
	log.Info("finish update", nil)
}
