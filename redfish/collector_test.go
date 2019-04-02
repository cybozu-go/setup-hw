package redfish

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func testCollect(t *testing.T) {
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

	expectedSet := []struct {
		name   string
		value  float64
		labels map[string]string
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
	collector.cache.set(dataMap)

	ch := make(chan prometheus.Metric)
	go collector.Collect(ch)

	for _, expected := range expectedSet {
		var actual prometheus.Metric
		select {
		case actual = <-ch:
		case <-time.After(1 * time.Second):
			t.Fatal("timeout to receive metric")
		}

		if !strings.Contains(actual.Desc().String(), `"`+expected.name+`"`) {
			t.Error("wrong metric name; expected:", expected.name, "actual in:", actual.Desc().String())
		}

		var metric dto.Metric
		err = actual.Write(&metric)
		if err != nil {
			t.Fatal(err)
		}

		if metric.Gauge == nil {
			t.Error("metric type is not Gauge:", expected.name)
		} else if metric.Gauge.GetValue() != expected.value {
			t.Error("wrong metric value; metric:", expected.name,
				"expected:", expected.value, "actual:", metric.Gauge.GetValue())
		}

		actualLabels := make(map[string]string)
		for _, label := range metric.Label {
			actualLabels[label.GetName()] = label.GetValue()
		}
		for k, v := range expected.labels {
			if _, ok := actualLabels[k]; !ok {
				t.Error("label not found; metric:", expected.name, "key:", k)
			} else if actualLabels[k] != v {
				t.Error("wrong label value; metric:", expected.name, "key:", k,
					"expected:", v, "actual:", actualLabels[k])
			}
		}
	}

	select {
	case <-ch:
		t.Error("collector returned extra metrics")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestCollector(t *testing.T) {
	t.Run("Collect", testCollect)
}
