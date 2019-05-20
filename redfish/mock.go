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
	defaultDummyData = `
[
  {
    "path": "/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1",
    "data": {
      "Status": {
        "Health": "OK"
      }
    }
  },
  {
    "path": "/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.2",
    "data": {
      "Status": {
        "Health": "OK"
      }
    }
  },
  {
    "path": "/redfish/v1/Systems/System.Embedded.1/Storage/AHCI.Slot.1-1",
    "data": {
      "Status": {
        "Health": "OK"
      }
    }
  },
  {
    "path": "/redfish/v1/Systems/System.Embedded.1/Storage/PCIeSSD.Slot.2-C",
    "data": {
      "Status": {
        "Health": "OK"
      }
    }
  },
  {
    "path": "/redfish/v1/Systems/System.Embedded.1/Storage/PCIeSSD.Slot.3-C",
    "data": {
      "Status": {
        "Health": "OK"
      }
    }
  }
]`
)

type dummyData struct {
	Path string      `json:"path"`
	Data interface{} `json:"data"`
}

type mockClient struct {
	filename    string
	defaultData dataMap
}

// NewMockClient create a mock client mock.
func NewMockClient(filename string) Client {
	return &mockClient{
		filename:    filename,
		defaultData: makeDataMap([]byte(defaultDummyData)),
	}
}

func makeDataMap(data []byte) dataMap {
	dataMap := make(dataMap)
	var dummyMetrics []dummyData

	if err := json.Unmarshal(data, &dummyMetrics); err != nil {
		log.Error("cannot unmarshal dummy data", map[string]interface{}{
			log.FnError: err,
		})
		return dataMap
	}

	for _, dummy := range dummyMetrics {
		container, err := gabs.Consume(dummy.Data)
		if err != nil {
			log.Error("failed to consume", map[string]interface{}{
				log.FnError: err,
			})
			continue
		}
		dataMap[dummy.Path] = container
	}
	return dataMap
}

func (c *mockClient) traverse(ctx context.Context) dataMap {
	cBytes, err := ioutil.ReadFile(c.filename)
	if err != nil {
		log.Error("cannot open dummy data file: "+c.filename, map[string]interface{}{
			log.FnError: err,
		})
		return c.defaultData
	}

	return makeDataMap(cBytes)
}
