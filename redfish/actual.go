package redfish

import (
	"context"
	"encoding/json"
	"os"
	"sort"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/gabs"
	prommodel "github.com/prometheus/client_model/go"
)

type metrics struct {
	name   string
	typ    prommodel.MetricType
	value  float64
	labels map[string]string
}

func (m metrics) key() string {
	keys := make([]string, 0, len(m.labels))
	for k := range m.labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	key := m.name + "{"
	for i, k := range keys {
		if i > 0 {
			key += ","
		}
		key += k + "=" + m.labels[k]
	}
	key += "}"
	return key
}

type actualData map[string]map[string]interface{}

var defaultActualData []byte

type actualClient struct {
	filename   string
	actualData dataMap
}

// NewMockClient create a mock client mock.
func NewActualClient(filename string) Client {
	return &actualClient{
		filename:   filename,
		actualData: makeActualMap([]byte(defaultActualData)),
	}
}

func makeActualMap(data []byte) dataMap {
	dataMap := make(dataMap)
	var actualMetrics actualData
	if err := json.Unmarshal(data, &actualMetrics); err != nil {
		log.Error("cannot unmarshal actual data", map[string]interface{}{
			log.FnError: err,
		})
		return dataMap
	}
	for key, value := range actualMetrics {
		container, err := gabs.Consume(value)
		if err != nil {
			log.Error("failed to consume", map[string]interface{}{
				log.FnError: err,
			})
			continue
		}
		dataMap[key] = container
	}
	return dataMap
}

func (c *actualClient) Traverse(ctx context.Context, rule *CollectRule) Collected {
	cBytes, err := os.ReadFile(c.filename)

	if err != nil {
		log.Error("cannot open dummy data file: "+c.filename, map[string]interface{}{
			log.FnError: err,
		})
		return Collected{data: c.actualData, rule: rule}
	}
	return Collected{data: makeActualMap(cBytes), rule: rule}
}

func (c *actualClient) GetVersion(ctx context.Context) (string, error) {
	return "1.0.0", nil
}
