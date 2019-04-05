package redfish

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/gabs"
)

func testUpdate(t *testing.T) {
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

	rule, err := ioutil.ReadFile("../testdata/redfish_metrics.yml")
	if err != nil {
		t.Fatal(err)
	}

	cc := &CollectorConfig{
		AddressConfig: &config.AddressConfig{IPv4: config.IPv4Config{Address: hostAndPort[0]}},
		Port:          hostAndPort[1],
		UserConfig:    &config.UserConfig{},
		Rule:          rule,
	}
	collector, err := NewCollector(cc)
	if err != nil {
		t.Fatal(err)
	}

	collector.Update(context.Background(), inputs[0].urlPath)
	v := collector.dataMap.Load()
	if v == nil {
		t.Fatal(errors.New("Update() did not store traversed data"))
	}
	dataMap := v.(dataMap)

	for _, input := range inputs {
		data, ok := dataMap[input.urlPath]
		if !ok {
			t.Error("path not traversed:", input.urlPath)
			continue
		}

		inputData, err := gabs.ParseJSONFile(input.filePath)
		if err != nil {
			t.Fatal(err)
		}

		if data.String() != inputData.String() {
			t.Error("wrong contents loaded:", input.urlPath,
				"\nexpected:", inputData.String(), "\nactual:", data.String())
			continue
		}
	}
	if len(dataMap) > len(inputs) {
		t.Error("extra path was traversed")
	}
}

func TestRedfish(t *testing.T) {
	t.Run("Update", testUpdate)
}
