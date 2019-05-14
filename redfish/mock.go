package redfish

import (
	"context"
	"encoding/json"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/gabs"
	"io/ioutil"
)

const (
	// DummyRedfishFile is the filename of dummy data for Redfish API.
	DummyRedfishFile = "/etc/neco/dummy_redfish_data.json"
)

type dummyData struct {
	Path string      `json:"path"`
	Data interface{} `json:"data"`
}

type mockClient struct {
	filename string
}

// NewMockClient create a mock client mock.
func NewMockClient(filename string) Client {
	return &mockClient{filename: filename}
}

func (c *mockClient) traverse(ctx context.Context) dataMap {
	dataMap := make(dataMap)
	var dummyMetrics []dummyData

	cBytes, err := ioutil.ReadFile(c.filename)
	if err != nil {
		log.Error("cannot open dummy data file: "+c.filename, map[string]interface{}{
			log.FnError: err,
		})
		return dataMap
	}

	if err := json.Unmarshal(cBytes, &dummyMetrics); err != nil {
		log.Error("cannot unmarshal dummy data file: "+c.filename, map[string]interface{}{
			log.FnError: err,
		})
		return dataMap
	}

	for _, dummy := range dummyMetrics {
		container, err := gabs.Consume(dummy.Data)
		if err != nil {
			log.Error("failed to consume: "+dummy.Path, map[string]interface{}{
				log.FnError: err,
			})
			continue
		}
		dataMap[dummy.Path] = container
	}
	return dataMap
}
