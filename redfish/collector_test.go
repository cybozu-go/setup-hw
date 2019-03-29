package redfish

import (
	"testing"
	"time"

	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// "address": "1.2.3.4",
// "netmask": "255.255.255.0",
// "gateway": "1.2.3.1"

func testCollect(t *testing.T) {
	ac := &config.AddressConfig{IPv4: config.IPv4Config{Address: "1.2.3.4"}}
	uc := &config.UserConfig{}
	ruleFile := "../testdata/test_redfish.yml"
	collector, err := NewRedfishCollector(ac, uc, ruleFile)
	if err != nil {
		t.Fatal(err)
	}
	chassis, err := gabs.ParseJSONFile("../testdata/redfish_chassis.json")
	if err != nil {
		t.Fatal(err)
	}
	dataMap := RedfishDataMap{"/redfish/v1/Chassis/System.Embedded.1": chassis}
	collector.cache.Set(dataMap)
	ch := make(chan prometheus.Metric)
	go collector.Collect(ch)
	var actual prometheus.Metric
	select {
	case actual = <-ch:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout to receive metric")
	}
	var metric dto.Metric
	err = actual.Write(&metric)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCollector(t *testing.T) {
	t.Run("Collect", testCollect)
}
