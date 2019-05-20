package redfish

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

type expected struct {
	matched bool
	name    string
	value   float64
	labels  map[string]string
}

func TestMockClient(t *testing.T) {
	t.Parallel()

	expectedSet := []*expected{
		{
			name:  "hw_processor_status_health",
			value: 0, // OK
			labels: map[string]string{
				"system":    "System.Embedded.1",
				"processor": "CPU.Socket.1",
			},
		},
		{
			name:  "hw_processor_status_health",
			value: 1, // Warning
			labels: map[string]string{
				"system":    "System.Embedded.1",
				"processor": "CPU.Socket.2",
			},
		},
		{
			name:  "hw_storage_controller_status_health",
			value: 0, // OK
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"controller": "AHCI.Slot.1-1",
			},
		},
		{
			name:  "hw_storage_controller_status_health",
			value: 1, // Warning
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"controller": "PCIeSSD.Slot.2-C",
			},
		},
		{
			name:  "hw_storage_controller_status_health",
			value: 2, // Critical
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"controller": "PCIeSSD.Slot.3-C",
			},
		},
	}

	rule, err := collectRule("../testdata/mock_collect.yml")
	if err != nil {
		t.Fatal(err)
	}

	client := NewMockClient("../testdata/mock_data.json")

	checkResult(t, rule, client, expectedSet)
}

func TestMockClientDefaultData(t *testing.T) {
	t.Parallel()

	expectedSet := []*expected{
		{
			name:  "hw_processor_status_health",
			value: 0, // OK
			labels: map[string]string{
				"system":    "System.Embedded.1",
				"processor": "CPU.Socket.1",
			},
		},
		{
			name:  "hw_processor_status_health",
			value: 0, // OK
			labels: map[string]string{
				"system":    "System.Embedded.1",
				"processor": "CPU.Socket.2",
			},
		},
		{
			name:  "hw_storage_controller_status_health",
			value: 0, // OK
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"controller": "AHCI.Slot.1-1",
			},
		},
		{
			name:  "hw_storage_controller_status_health",
			value: 0, // OK
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"controller": "PCIeSSD.Slot.2-C",
			},
		},
		{
			name:  "hw_storage_controller_status_health",
			value: 0, // OK
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"controller": "PCIeSSD.Slot.3-C",
			},
		},
	}

	client := NewMockClient("../testdata/no_exist_file")

	checkResult(t, Rules["qemu.yml"], client, expectedSet)
}

func checkResult(t *testing.T, rule *CollectRule, client Client, expectedSet []*expected) {
	collector, err := NewCollector(rule, client)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	collector.Update(ctx)

	registry := prometheus.NewPedanticRegistry()
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
