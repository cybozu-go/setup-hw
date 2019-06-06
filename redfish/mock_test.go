package redfish

import (
	"context"
	"math"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	prommodel "github.com/prometheus/client_model/go"
)

type expected struct {
	matched bool
	name    string
	typ     prommodel.MetricType
	value   float64
	labels  map[string]string
}

func TestMockClient(t *testing.T) {
	t.Parallel()

	expectedSet := []*expected{
		{
			name:  "hw_processor_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0, // OK
			labels: map[string]string{
				"system":    "System.Embedded.1",
				"processor": "CPU.Socket.1",
			},
		},
		{
			name:  "hw_processor_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 1, // Warning
			labels: map[string]string{
				"system":    "System.Embedded.1",
				"processor": "CPU.Socket.2",
			},
		},
		{
			name:  "hw_storage_controller_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0, // OK
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"controller": "AHCI.Slot.1-1",
			},
		},
		{
			name:  "hw_storage_controller_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 1, // Warning
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"controller": "PCIeSSD.Slot.2-C",
			},
		},
		{
			name:  "hw_storage_controller_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 2, // Critical
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"controller": "PCIeSSD.Slot.3-C",
			},
		},
		{
			name:   "hw_last_update",
			typ:    prommodel.MetricType_COUNTER,
			value:  math.NaN(), // don't care
			labels: map[string]string{},
		},
		{
			name:   "hw_last_update_duration_minutes",
			typ:    prommodel.MetricType_GAUGE,
			value:  math.NaN(), // don't care
			labels: map[string]string{},
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
			typ:   prommodel.MetricType_GAUGE,
			value: 0, // OK
			labels: map[string]string{
				"system":    "System.Embedded.1",
				"processor": "CPU.Socket.1",
			},
		},
		{
			name:  "hw_processor_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0, // OK
			labels: map[string]string{
				"system":    "System.Embedded.1",
				"processor": "CPU.Socket.2",
			},
		},
		{
			name:  "hw_storage_controller_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0, // OK
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"controller": "AHCI.Slot.1-1",
			},
		},
		{
			name:  "hw_storage_controller_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0, // OK
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"controller": "PCIeSSD.Slot.2-C",
			},
		},
		{
			name:  "hw_storage_controller_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0, // OK
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"controller": "PCIeSSD.Slot.3-C",
			},
		},
		{
			name:   "hw_last_update",
			typ:    prommodel.MetricType_COUNTER,
			value:  math.NaN(), // don't care
			labels: map[string]string{},
		},
		{
			name:   "hw_last_update_duration_minutes",
			typ:    prommodel.MetricType_GAUGE,
			value:  math.NaN(), // don't care
			labels: map[string]string{},
		},
	}

	client := NewMockClient("../testdata/no_exist_file")

	checkResult(t, Rules["qemu.yml"], client, expectedSet)
}

func checkResult(t *testing.T, rule *CollectRule, client Client, expectedSet []*expected) {
	collector, err := NewCollector(func() (*CollectRule, error) {
		return rule, nil
	}, client)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	collector.Update(ctx)

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
			actualLabels := make(map[string]string)
			for _, label := range actual.GetLabel() {
				actualLabels[label.GetName()] = label.GetValue()
			}

			for _, expected := range expectedSet {
				if expected.matched {
					continue
				}
				if expected.typ != *metricFamily.Type {
					continue
				}
				var val float64
				switch *metricFamily.Type {
				case prommodel.MetricType_GAUGE:
					val = actual.GetGauge().GetValue()
				case prommodel.MetricType_COUNTER:
					val = actual.GetCounter().GetValue()
				default:
					t.Fatalf("unknown type: ")
				}
				if actualMetricName == expected.name &&
					(math.IsNaN(expected.value) || val == expected.value) &&
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
