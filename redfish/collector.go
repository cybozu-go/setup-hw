package redfish

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/prometheus/client_golang/prometheus"
	yaml "gopkg.in/yaml.v2"
)

// Namespace is the first part of the metrics name.
const Namespace = "hw"

type RedfishCollector struct {
	rules    []ConvertRule
	rfclient *Redfish
	cache    *Cache
}

func NewRedfishCollector(ac *config.AddressConfig, uc *config.UserConfig, ruleFile string) (*RedfishCollector, error) {
	cache := &Cache{dataMap: make(RedfishDataMap)}
	rfclient, err := NewRedfish(ac, uc, cache)
	if err != nil {
		return nil, err
	}

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

	return &RedfishCollector{rules: rules, rfclient: rfclient, cache: cache}, nil
}

func (c RedfishCollector) Describe(ch chan<- *prometheus.Desc) {
}

func (c RedfishCollector) Collect(ch chan<- prometheus.Metric) {
	dataMap := c.cache.Get()

	for _, rule := range c.rules {
		for path, parsedJSON := range dataMap {
			matched, pathLabels := matchPath(rule.Path, path)
			if !matched {
				continue
			}

			for _, propertyRule := range rule.Rules {
				matchedProperties := matchPointer(propertyRule.Pointer, parsedJSON)
				for _, matched := range matchedProperties {
					value := propertyRule.Converter(matched.property)
					labels := make(prometheus.Labels)
					for k, v := range pathLabels {
						labels[k] = v
					}
					for k, v := range matched.indexes {
						labels[k] = strconv.Itoa(v)
					}
					desc := prometheus.NewDesc(
						prometheus.BuildFQName(Namespace, "", propertyRule.Name),
						propertyRule.Description, nil, labels)
					ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value)
				}
			}
		}
	}
}

func (c *RedfishCollector) Update(ctx context.Context, rootPath string) {
	c.rfclient.Update(ctx, rootPath)
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
		if ruleElements[i][0] == '{' && ruleElements[i][ln-1] == '}' {
			labelName := ruleElements[i][1 : ln-1]
			labels[labelName] = pathElements[i]
		} else if ruleElements[i] != pathElements[i] {
			return false, nil
		}
	}

	return true, labels
}

type RedfishDataMap map[string]*gabs.Container

type Cache struct {
	dataMap RedfishDataMap
	mux     sync.Mutex
}

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

func matchPointer(pointer string, parsedJSON *gabs.Container) []matchedProperty {
	return matchPointerAux(pointer, parsedJSON, pointer)
}

func matchPointerAux(pointer string, parsedJSON *gabs.Container, rootPointer string) []matchedProperty {
	if pointer == "" {
		return []matchedProperty{
			matchedProperty{
				property: parsedJSON.Data(),
				indexes:  make(map[string]int),
			},
		}
	}

	if pointer[0] != '/' {
		log.Warn("pointer must begin with '/'", map[string]interface{}{
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
				"pointer": rootPointer,
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
			"pointer": rootPointer,
		})
		return nil
	}

	children, err := v.Children()
	if err != nil {
		log.Warn("index pattern is used, but parent is not array", map[string]interface{}{
			"pointer": rootPointer,
		})
		return nil
	}

	var result []matchedProperty
	for i, child := range children {
		ms := matchPointerAux(remainder, child, rootPointer)
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
		if t[0] == '{' && t[len(t)-1] == '}' {
			matched = true
			subpath = "/" + strings.Join(ts[0:i], "/")
			index = t[1 : len(t)-1]
			if i != len(ts)-1 {
				remainder = "/" + strings.Join(ts[i+1:len(ts)], "/")
			}
			return
		}
	}
	return false, "", "", ""
}
