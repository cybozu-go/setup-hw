package redfish

import (
	"context"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/prometheus/client_golang/prometheus"
	yaml "gopkg.in/yaml.v2"
)

const namespace = "hw"

type dataMap map[string]*gabs.Container

// Collector implements prometheus.Collector interface.
type Collector struct {
	rule    *CollectRule
	client  *client
	dataMap atomic.Value
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
	rule := new(CollectRule)
	err := yaml.Unmarshal(cc.Rule, rule)
	if err != nil {
		return nil, err
	}

	if err := rule.Validate(); err != nil {
		return nil, err
	}

	if err := rule.Compile(); err != nil {
		return nil, err
	}

	client, err := newClient(cc, rule.TraverseRule)
	if err != nil {
		return nil, err
	}

	return &Collector{rule: rule, client: client}, nil
}

// Describe sends descriptions of metrics.
func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, rule := range c.rule.MetricRules {
		for _, propertyRule := range rule.PropertyRules {
			ch <- propertyRule.desc
		}
	}
}

// Collect sends metrics collected from BMC via Redfish.
func (c Collector) Collect(ch chan<- prometheus.Metric) {
	v := c.dataMap.Load()
	if v == nil {
		return
	}
	dataMap := v.(dataMap)

	for _, rule := range c.rule.MetricRules {
		for path, parsedJSON := range dataMap {
			matched, pathLabelValues := matchPath(rule.Path, path)
			if !matched {
				continue
			}

			for _, propertyRule := range rule.PropertyRules {
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

					var labelValues []string
					labelValues = append(labelValues, pathLabelValues...)
					for _, v := range matched.indexes {
						labelValues = append(labelValues, strconv.Itoa(v))
					}

					m, err := prometheus.NewConstMetric(propertyRule.desc, prometheus.GaugeValue, value, labelValues...)
					if err != nil {
						log.Warn("failed to create metric", map[string]interface{}{
							"path":      path,
							"pointer":   propertyRule.Pointer,
							"name":      propertyRule.Name,
							"value":     matched.property,
							log.FnError: err,
						})
						continue
					}

					ch <- m
				}
			}
		}
	}
}

// Update collects metrics from BMCs via Redfish.
func (c *Collector) Update(ctx context.Context) {
	dataMap := c.client.traverse(ctx)
	c.dataMap.Store(dataMap)
}

func getLabelName(elem string) (string, bool) {
	ln := len(elem)
	if ln >= 3 && elem[0] == '{' && elem[ln-1] == '}' {
		return elem[1 : ln-1], true
	}
	return "", false
}

func getLabelNamesInPath(path string) []string {
	var labelNames []string
	for _, elem := range strings.Split(path, "/") {
		if name, ok := getLabelName(elem); ok {
			labelNames = append(labelNames, name)
		}
	}
	return labelNames
}

func matchPath(rulePath, path string) (bool, []string) {
	ruleElements := strings.Split(rulePath, "/")
	pathElements := strings.Split(path, "/")

	if len(ruleElements) != len(pathElements) {
		return false, nil
	}

	var labelValues []string
	for i := 0; i < len(ruleElements); i++ {
		if _, ok := getLabelName(ruleElements[i]); ok {
			labelValues = append(labelValues, pathElements[i])
		} else if ruleElements[i] != pathElements[i] {
			return false, nil
		}
	}

	return true, labelValues
}

type matchedProperty struct {
	property interface{}
	indexes  []int
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
				indexes:  nil,
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

	hasIndexPattern, subpath, remainder := splitPointer(pointer)
	if !hasIndexPattern {
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
				indexes:  nil,
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
			m.indexes = append([]int{i}, m.indexes...)
			result = append(result, m)
		}
	}

	return result
}

func splitPointer(pointer string) (hasIndexPattern bool, subpath, remainder string) {
	ts := strings.Split(pointer, "/")
	for i, t := range ts {
		if _, ok := getLabelName(t); ok {
			hasIndexPattern = true
			subpath = strings.Join(ts[0:i], "/")
			if i != len(ts)-1 {
				remainder = "/" + strings.Join(ts[i+1:], "/")
			}
			return
		}
	}
	return false, "", ""
}
