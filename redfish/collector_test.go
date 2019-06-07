package redfish

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/ghodss/yaml"
	"github.com/prometheus/client_golang/prometheus"
	prommodel "github.com/prometheus/client_model/go"
)

func collectRule(filename string) (*CollectRule, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	rule := new(CollectRule)
	yaml.Unmarshal(data, rule)
	if err := rule.Validate(); err != nil {
		return nil, err
	}
	if err := rule.Compile(); err != nil {
		return nil, err
	}
	return rule, nil
}

func clientConfig() (*ClientConfig, error) {
	rule, err := collectRule("../testdata/redfish_collect.yml")
	if err != nil {
		return nil, err
	}

	return &ClientConfig{
		AddressConfig: &config.AddressConfig{IPv4: config.IPv4Config{Address: "1.2.3.4"}},
		UserConfig:    &config.UserConfig{},
		Rule:          rule,
	}, nil
}

func testDescribe(t *testing.T) {
	t.Parallel()

	expectedList := []struct {
		name       string
		help       string
		labelNames []string
	}{
		{
			name:       "hw_last_update",
			labelNames: []string{},
		},
		{
			name:       "hw_last_update_duration_minutes",
			labelNames: []string{},
		},
	}

	cc, err := clientConfig()
	if err != nil {
		t.Fatal(err)
	}

	client, err := NewRedfishClient(cc)
	if err != nil {
		t.Fatal(err)
	}

	collector, err := NewCollector(func(context.Context) (*CollectRule, error) {
		return cc.Rule, nil
	}, client)
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan *prometheus.Desc)
	go collector.Describe(ch)

	for _, expected := range expectedList {
		var actual *prometheus.Desc
		select {
		case actual = <-ch:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout to receive description")
		}

		// We cannot access fields of prometheus.Desc, so take String() for comparison.
		if !strings.Contains(actual.String(), `"`+expected.name+`"`) {
			t.Error("wrong name in description; expected:", expected.name, "actual in:", actual.String())
		}

		if !strings.Contains(actual.String(), `"`+expected.help+`"`) {
			t.Error("wrong help in description; expected:", expected.help, "actual in:", actual.String())
		}

		if !strings.Contains(actual.String(), fmt.Sprint(expected.labelNames)) {
			t.Error("wrong variable label names in description; expected:", expected.labelNames, "actual in:", actual.String())
		}
	}

	select {
	case <-ch:
		t.Error("collector returned extra descriptions")
	case <-time.After(100 * time.Millisecond):
	}
}

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
			urlPath:  "/redfish/v1/Chassis/System.Embedded.1/Blocks/0",
			filePath: "../testdata/redfish_block.json",
		},
	}

	expectedSet := []*struct {
		matched bool
		name    string
		typ     prommodel.MetricType
		value   float64
		labels  map[string]string
	}{
		{
			name:  "hw_chassis_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0, // OK
			labels: map[string]string{
				"chassis": "System.Embedded.1",
			},
		},
		{
			name:  "hw_chassis_sub_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 1, // Warning
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"sub":     "0",
			},
		},
		{
			name:  "hw_chassis_sub_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 2, // Critical
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"sub":     "2",
			},
		},
		{
			name:  "hw_block_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 1, // Warning
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"block":   "0",
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

	cc, err := clientConfig()
	if err != nil {
		t.Fatal(err)
	}

	client, err := NewRedfishClient(cc)
	if err != nil {
		t.Fatal(err)
	}

	collector, err := NewCollector(func(context.Context) (*CollectRule, error) {
		return cc.Rule, nil
	}, client)
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
	collector.collected.Store(Collected{data: dataMap, rule: cc.Rule})

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

func testUpdate(t *testing.T) {
	t.Parallel()

	inputs := []struct {
		urlPath  string
		filePath string
		needed   bool
	}{
		{
			urlPath:  "/redfish/v1/Chassis/System.Embedded.1",
			filePath: "../testdata/redfish_chassis.json",
			needed:   true,
		},
		{
			urlPath:  "/redfish/v1/Chassis/System.Embedded.1/Blocks/0",
			filePath: "../testdata/redfish_block.json",
			needed:   true,
		},
		{
			urlPath:  "/redfish/v1/Chassis/System.Embedded.1/Trashes/0",
			filePath: "../testdata/redfish_trash.json",
			needed:   false,
		},
	}

	mux := http.NewServeMux()
	for _, input := range inputs {
		input := input
		mux.HandleFunc(input.urlPath, func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, input.filePath)
		})
	}
	ts := httptest.NewTLSServer(mux)
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	hostAndPort := strings.Split(u.Host, ":")
	if len(hostAndPort) != 2 {
		t.Fatal(errors.New("httptest.NewTLSServer() returned URL with host and/or port omitted"))
	}

	cc, err := clientConfig()
	if err != nil {
		t.Fatal(err)
	}

	cc.AddressConfig = &config.AddressConfig{IPv4: config.IPv4Config{Address: hostAndPort[0]}}
	cc.Port = hostAndPort[1]

	client, err := NewRedfishClient(cc)
	if err != nil {
		t.Fatal(err)
	}

	collector, err := NewCollector(func(context.Context) (*CollectRule, error) {
		return cc.Rule, nil
	}, client)
	if err != nil {
		t.Fatal(err)
	}

	collector.Update(context.Background())
	v := collector.collected.Load()
	if v == nil {
		t.Fatal(errors.New("Update() did not store traversed data"))
	}
	cl := v.(Collected)

	for _, input := range inputs {
		if !input.needed {
			continue
		}

		data, ok := cl.data[input.urlPath]
		if !ok {
			t.Error("path was not traversed:", input.urlPath)
			continue
		}

		inputData, err := gabs.ParseJSONFile(input.filePath)
		if err != nil {
			t.Fatal(err)
		}

		if data.String() != inputData.String() {
			t.Error("wrong contents were loaded:", input.urlPath,
				"\nexpected:", inputData.String(), "\nactual:", data.String())
			continue
		}
	}

ActualLoop:
	for path := range cl.data {
		for _, input := range inputs {
			if path == input.urlPath && input.needed {
				continue ActualLoop
			}
		}
		t.Error("extra path was traversed:", path)
	}
}

func TestCollector(t *testing.T) {
	t.Run("Describe", testDescribe)
	t.Run("Collect", testCollect)
	t.Run("Update", testUpdate)
}
