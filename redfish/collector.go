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
	if pointer == "" {
		return []matchedProperty{
			matchedProperty{
				property: parsedJSON.Data(),
				indexes:  make(map[string]int),
			},
		}
	}
	if pointer[0] != '/' {
		log.Warn("wrong pointer path", map[string]interface{}{
			"pointer": pointer,
		})
		return nil
	}
	matched, subpath, index, remainder := splitPointer(pointer)
	if !matched {
		p := strings.ReplaceAll(pointer[1:], "/", ".")
		v := parsedJSON.Path(p)
		if v == nil {
			log.Warn("cannot find pointed value", map[string]interface{}{
				"pointer": pointer,
			})
			return nil
		}
		return []matchedProperty{
			matchedProperty{
				property: v.Data(),
				indexes:  make(map[string]int),
			},
		}
	}
	p := strings.ReplaceAll(subpath[1:], "/", ".")
	v := parsedJSON.Path(p)
	if v == nil {
		log.Warn("cannot find pointed value", map[string]interface{}{
			"pointer": pointer,
		})
		return nil
	}
	// Check if v is slice before using Children() because Children() works also for map.
	_, ok := v.Data().([]interface{})
	if !ok {
		log.Warn("array not found", map[string]interface{}{
			"pointer": pointer,
		})
		return nil
	}
	var result []matchedProperty
	children, err := v.Children()
	if err != nil {
		log.Warn("get gabs.Children() failed", map[string]interface{}{
			"pointer": pointer,
		})
		return nil
	}
	for i, child := range children {
		ms := findJSONPointerPattern(child, remainder)
		for _, m := range ms {
			m.indexes[index] = i
			result = append(result, m)
		}
	}
	return result
}

func splitPointer(pointer string) (matched bool, subpath, index, remainder string) {
	ts := strings.SplitN(pointer, "[", 2)
	if len(ts) != 2 {
		return false, "", "", ""
	}
	subpath = ts[0]

	ts = strings.SplitN(ts[1], "]", 2)
	if len(ts) != 2 {
		return false, "", "", ""
	}
	index = ts[0]
	remainder = ts[1]
	matched = true
	return
}
