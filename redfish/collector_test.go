package redfish

import (
	"io/ioutil"
	"testing"

	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/prometheus/client_golang/prometheus"
)

func testCollect(t *testing.T) {
	t.Parallel()

	inputs := []struct {
		urlPath  string
		filePath string
	}{
		{
			urlPath:  "/redfish/v1/Chassis/System.Embedded.1",
			filePath: "../testdata/redfish_chassis.json",
		},
		{
			urlPath:  "/redfish/v1/Chassis/System.Embedded.1/Block/0",
			filePath: "../testdata/redfish_block.json",
		},
	}

	expectedSet := []*struct {
		matched bool
		name    string
		value   float64
		labels  map[string]string
	}{
		{
			name:  "hw_chassis_status_health",
			value: 0, // OK
			labels: map[string]string{
				"chassis": "System.Embedded.1",
			},
		},
		{
			name:  "hw_chassis_sub_status_health",
			value: 1, // Warning
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"sub":     "0",
			},
		},
		{
			name:  "hw_chassis_sub_status_health",
			value: 2, // Critical
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"sub":     "1",
			},
		},
		{
			name:  "hw_block_status_health",
			value: 1, // Warning
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"block":   "0",
			},
		},
	}

	rule, err := ioutil.ReadFile("../testdata/redfish_metrics.yml")
	if err != nil {
		t.Fatal(err)
	}

	cc := &CollectorConfig{
		AddressConfig: &config.AddressConfig{IPv4: config.IPv4Config{Address: "1.2.3.4"}},
		UserConfig:    &config.UserConfig{},
		Rule:          rule,
	}
	collector, err := NewCollector(cc)
	if err != nil {
		t.Fatal(err)
	}

	dataMap := make(dataMap)
	for _, input := range inputs {
		data, err := gabs.ParseJSONFile(input.filePath)
		if err != nil {
			t.Fatal(err)
		}
		dataMap[input.urlPath] = data
	}
	collector.dataMap.Store(dataMap)

	registry := prometheus.NewRegistry()
	err = registry.Register(collector)
	if err != nil {
		t.Fatal(err)
	}

	// This calls collector.Collect() internally.
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}

	for _, metricFamily := range metricFamilies {
		actualMetricName := metricFamily.GetName()
	ActualLoop:
		for _, actual := range metricFamily.GetMetric() {
			if actual.GetGauge() == nil {
				t.Error("metric type is not Gauge:", actualMetricName)
				continue
			}

			actualLabels := make(map[string]string)
			for _, label := range actual.GetLabel() {
				actualLabels[label.GetName()] = label.GetValue()
			}

			for _, expected := range expectedSet {
				if expected.matched {
					continue
				}
				if actualMetricName == expected.name &&
					actual.GetGauge().GetValue() == expected.value &&
					matchLabels(actualLabels, expected.labels) {
					expected.matched = true
					continue ActualLoop
				}
			}
			t.Error("unexpected metric; name:", actualMetricName, "value:", actual.GetGauge().GetValue(), "labels:", actualLabels)
		}
	}

	for _, expected := range expectedSet {
		if !expected.matched {
			t.Error("expected but not returned; name:", expected.name, "value:", expected.value, "labels:", expected.labels)
		}
	}
}

func matchLabels(actual, expected map[string]string) bool {
	if len(actual) != len(expected) {
		return false
	}

	for k, v := range expected {
		if act, ok := actual[k]; !ok || act != v {
			return false
		}
	}

	return true
}

func TestCollector(t *testing.T) {
	t.Run("Collect", testCollect)
}
