package collector

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type chassisCollector struct {
	current *prometheus.Desc
}

func NewChassisCollector() Collector {
	return &chassisCollector{}
}

func (c *chassisCollector) Collect(ch chan<- prometheus.Metric) error {
	values := Metrics.Get("chassis")
	for _, value := range values {
		float, err := strconv.ParseFloat(value.Value, 64)
		if err != nil {
			return err
		}
		c.current = prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", value.Name),
			"Overall status of chassis components.",
			nil, value.Labels)
		ch <- prometheus.MustNewConstMetric(
			c.current, prometheus.GaugeValue, float)
	}
	return nil
}
