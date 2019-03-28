package redfish

import (
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/Jeffail/gabs"
	"github.com/cybozu-go/log"
	"github.com/prometheus/client_golang/prometheus"
	yaml "gopkg.in/yaml.v2"
)

func init() {
	cache = &Cache{dataMap: make(RedfishDataMap)}
}

// Namespace is the first part of the metrics name.
const Namespace = "hw"

type RedfishCollector struct {
	rules []ConvertRule
}

func NewRedfishCollector(ruleFile string) (*RedfishCollector, error) {
	f, err := os.Open(ruleFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var rules []ConvertRule
	err = yaml.NewDecoder(f).Decode(&rules)
	if err != nil {
		return nil, err
	}

	return &RedfishCollector{rules: rules}, nil
}

func (c RedfishCollector) Describe(ch chan<- *prometheus.Desc) {
}

func (c RedfishCollector) Collect(ch chan<- prometheus.Metric) {
	dataMap := cache.Get()

	for _, rule := range c.rules {
		// rule.Path:    "/redfish/v1/Chassis/{chassis}"
		// pathPattern: "^/redfish/v1/Chassis/(?P<chassis>[^/]+)$"
		pathPattern := "^" + strings.ReplaceAll(strings.ReplaceAll(rule.Path, "{", "(?P<"), "}", ">[^/]+") + "$"
		pathRegExp, err := regexp.Compile(pathPattern)
		if err != nil {
			log.Warn("wrong path pattern in metrics rule", map[string]interface{}{
				"path":      rule.Path,
				log.FnError: err,
			})
			continue
		}

		for path, parsedJSON := range dataMap {
			if !pathRegExp.MatchString(path) {
				continue
			}

			for _, propertyRule := range rule.Rules {
				matchedProperties := findJSONPointerPattern(parsedJSON, propertyRule.Pointer)
				for _, matched := range matchedProperties {
					value := propertyRule.Converter(matched.property)
					labels := make(prometheus.Labels) //TODO
					desc := prometheus.NewDesc(
						prometheus.BuildFQName(Namespace, "", propertyRule.Name),
						propertyRule.Description, nil, labels)
					ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value)
				}
			}
		}
	}
}

type RedfishDataMap map[string]*gabs.Container

type Cache struct {
	dataMap RedfishDataMap
	mux     sync.Mutex
}

var cache *Cache

func (c *Cache) Set(dataMap RedfishDataMap) {
	c.mux.Lock()
	c.dataMap = dataMap
	c.mux.Unlock()
}

func (c *Cache) Get() RedfishDataMap {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.dataMap
}

type matchedProperty struct {
	property interface{}
	indexes  map[string]int
}

func findJSONPointerPattern(parsedJSON *gabs.Container, pointer string) []matchedProperty {
	return nil
	// "/A/B" => "A.B"
	// "/A[{id}]/B" => "/A" -> (list processing) -> "/B"
	//
}
