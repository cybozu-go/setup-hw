package redfish

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/prometheus/client_golang/prometheus"
	yaml "gopkg.in/yaml.v2"
)

const namespace = "hw"

// Collector implements prometheus.Collector interface.
type Collector struct {
	rules  []convertRule
	client *client
	cache  *cache
}

// CollectorConfig is a set of configurations for Collector.
type CollectorConfig struct {
	AddressConfig *config.AddressConfig
	Port          string
	UserConfig    *config.UserConfig
	Rule          []byte
}

// NewCollector returns a new instance of Collector.
func NewCollector(cc *CollectorConfig) (*Collector, error) {
	cache := &cache{dataMap: make(dataMap)}
	client, err := newClient(cc, cache)
	if err != nil {
		return nil, err
	}

	var rules []convertRule
	err = yaml.Unmarshal(cc.Rule, &rules)
	if err != nil {
		return nil, err
	}

	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			return nil, err
		}
	}

	return &Collector{rules: rules, client: client, cache: cache}, nil
}

// Describe does nothing for now.
func (c Collector) Describe(ch chan<- *prometheus.Desc) {
}

// Collect sends metrics collected from BMC via Redfish.
func (c Collector) Collect(ch chan<- prometheus.Metric) {
	dataMap := c.cache.get()

	for _, rule := range c.rules {
		for path, parsedJSON := range dataMap {
			matched, pathLabels := matchPath(rule.Path, path)
			if !matched {
				continue
			}

			for _, propertyRule := range rule.Rules {
				matchedProperties := matchPointer(propertyRule.Pointer, parsedJSON, path)
				for _, matched := range matchedProperties {
					value, err := propertyRule.Converter(matched.property)
					if err != nil {
						log.Warn("failed to interpret Redfish data as metric", map[string]interface{}{
							"path":      path,
							"pointer":   propertyRule.Pointer,
							"name":      propertyRule.Name,
							"value":     matched.property,
							log.FnError: err,
						})
						continue
					}
					labels := make(prometheus.Labels)
					for k, v := range pathLabels {
						labels[k] = v
					}
					for k, v := range matched.indexes {
						labels[k] = strconv.Itoa(v)
					}
					desc := prometheus.NewDesc(
						prometheus.BuildFQName(namespace, "", propertyRule.Name),
						propertyRule.Description, nil, labels)
					ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value)
				}
			}
		}
	}
}

// Update collects metrics from BMCs via Redfish.
func (c *Collector) Update(ctx context.Context, rootPath string) {
	c.client.update(ctx, rootPath)
}

func matchPath(rulePath, path string) (bool, prometheus.Labels) {
	ruleElements := strings.Split(rulePath, "/")
	pathElements := strings.Split(path, "/")

	if len(ruleElements) != len(pathElements) {
		return false, nil
	}

	labels := make(prometheus.Labels)
	for i := 0; i < len(ruleElements); i++ {
		ln := len(ruleElements[i])
		if ln >= 2 && ruleElements[i][0] == '{' && ruleElements[i][ln-1] == '}' {
			labelName := ruleElements[i][1 : ln-1]
			labels[labelName] = pathElements[i]
		} else if ruleElements[i] != pathElements[i] {
			return false, nil
		}
	}

	return true, labels
}

type dataMap map[string]*gabs.Container

type cache struct {
	dataMap dataMap
	mux     sync.Mutex
}

func (c *cache) set(dataMap dataMap) {
	c.mux.Lock()
	c.dataMap = dataMap
	c.mux.Unlock()
}

func (c *cache) get() dataMap {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.dataMap
}

type matchedProperty struct {
	property interface{}
	indexes  map[string]int
}

func matchPointer(pointer string, parsedJSON *gabs.Container, path string) []matchedProperty {
	// path is for logging
	return matchPointerAux(pointer, parsedJSON, path, pointer)
}

func matchPointerAux(pointer string, parsedJSON *gabs.Container, path, rootPointer string) []matchedProperty {
	// path and rootPointer are for logging
	if pointer == "" {
		return []matchedProperty{
			{
				property: parsedJSON.Data(),
				indexes:  make(map[string]int),
			},
		}
	}

	if pointer[0] != '/' {
		log.Warn("pointer must begin with '/'", map[string]interface{}{
			"path":    path,
			"pointer": rootPointer,
		})
		return nil
	}

	matched, subpath, index, remainder := splitPointer(pointer)
	if !matched {
		p := strings.ReplaceAll(pointer[1:], "/", ".")
		v := parsedJSON.Path(p)
		if v == nil {
			log.Warn("cannot find pointed value", map[string]interface{}{
				"path":    path,
				"pointer": rootPointer,
			})
			return nil
		}
		return []matchedProperty{
			{
				property: v.Data(),
				indexes:  make(map[string]int),
			},
		}
	}

	p := strings.ReplaceAll(subpath[1:], "/", ".")
	v := parsedJSON.Path(p)
	if v == nil {
		log.Warn("cannot find pointed value", map[string]interface{}{
			"path":    path,
			"pointer": rootPointer,
		})
		return nil
	}

	children, err := v.Children()
	if err != nil {
		log.Warn("index pattern is used, but parent is not array", map[string]interface{}{
			"path":    path,
			"pointer": rootPointer,
		})
		return nil
	}

	var result []matchedProperty
	for i, child := range children {
		ms := matchPointerAux(remainder, child, path, rootPointer)
		for _, m := range ms {
			m.indexes[index] = i
			result = append(result, m)
		}
	}

	return result
}

func splitPointer(pointer string) (matched bool, subpath, index, remainder string) {
	ts := strings.Split(pointer, "/")
	for i, t := range ts {
		if len(t) >= 2 && t[0] == '{' && t[len(t)-1] == '}' {
			matched = true
			subpath = strings.Join(ts[0:i], "/")
			index = t[1 : len(t)-1]
			if i != len(ts)-1 {
				remainder = "/" + strings.Join(ts[i+1:], "/")
			}
			return
		}
	}
	return false, "", "", ""
}
