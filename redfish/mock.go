package redfish

import (
	"context"
	"github.com/cybozu-go/setup-hw/gabs"
)

type mockClient struct {
}

// NewMockClient create a mock client mock.
func NewMockClient() Client {
	return &mockClient{}
}

func (c *mockClient) traverse(ctx context.Context) dataMap {
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

	dataMap := make(dataMap)
	for _, input := range inputs {
		data, err := gabs.ParseJSONFile(input.filePath)
		if err != nil {
			continue
		}
		dataMap[input.urlPath] = data
	}
	return dataMap
}
