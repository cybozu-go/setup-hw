package redfish

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/prometheus/client_golang/prometheus"
)

// CollectRule is a set of rules of traversing and converting Redfish data.
type CollectRule struct {
	TraverseRule traverseRule  `json:"Traverse" yaml:"Traverse"`
	MetricRules  []*metricRule `json:"Metrics" yaml:"Metrics"`
}

type traverseRule struct {
	Root          string   `json:"Root" yaml:"Root"`
	ExcludeRules  []string `json:"Excludes" yaml:"Excludes"`
	excludeRegexp *regexp.Regexp
}

type metricRule struct {
	Path          string          `json:"Path" yaml:"Path"`
	PropertyRules []*propertyRule `json:"Properties" yaml:"Properties"`
}

type propertyRule struct {
	Pointer   string `json:"Pointer" yaml:"Pointer"`
	Name      string `json:"Name" yaml:"Name"`
	Help      string `json:"Help" yaml:"Help"`
	Type      string `json:"Type" yaml:"Type"`
	converter converter
	desc      *prometheus.Desc
}

type matchedProperty struct {
	value   float64
	indexes []string
}

// Validate checks CollectRule and its Descendants.
func (cr CollectRule) Validate() error {
	if err := cr.TraverseRule.validate(); err != nil {
		return err
	}

	for _, metricRule := range cr.MetricRules {
		if err := metricRule.validate(); err != nil {
			return err
		}
	}

	return nil
}

// Compile fills private fields of CollectRule and its descendants.
func (cr *CollectRule) Compile() error {
	if err := cr.TraverseRule.compile(); err != nil {
		return err
	}

	for _, metricRule := range cr.MetricRules {
		if err := metricRule.compile(); err != nil {
			return err
		}
	}

	return nil
}

func (tr traverseRule) validate() error {
	if tr.Root == "" {
		return errors.New("`Root` is mandatory for traverse rule")
	}

	return nil
}

func (tr *traverseRule) compile() error {
	if len(tr.ExcludeRules) > 0 {
		excludes := strings.Join(tr.ExcludeRules, "|")
		r, err := regexp.Compile(excludes)
		if err != nil {
			return err
		}
		tr.excludeRegexp = r
	}

	return nil
}

func (mr metricRule) validate() error {
	if mr.Path == "" {
		return errors.New("`Path` is mandatory for metric rule")
	}

	for _, propertyRule := range mr.PropertyRules {
		if err := propertyRule.validate(); err != nil {
			return err
		}
	}

	return nil
}

func (mr *metricRule) compile() error {
	labelNames := getLabelNamesInPath(mr.Path)

	for _, propertyRule := range mr.PropertyRules {
		if err := propertyRule.compile(labelNames); err != nil {
			return err
		}
	}

	return nil
}

func (mr metricRule) matchDataMap(dataMap dataMap) []prometheus.Metric {
	var results []prometheus.Metric

	for path, parsedJSON := range dataMap {
		if matched, pathLabelValues := mr.matchPath(path); matched {
			metrics := mr.matchData(parsedJSON, pathLabelValues, path)
			results = append(results, metrics...)
		}
	}

	return results
}

func (mr metricRule) matchPath(path string) (bool, []string) {
	ruleElements := strings.Split(mr.Path, "/")
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

func (mr metricRule) matchData(parsedJSON *gabs.Container, pathLabelValues []string, loggedPath string) []prometheus.Metric {
	var results []prometheus.Metric

	for _, propertyRule := range mr.PropertyRules {
		metrics := propertyRule.matchPointer(parsedJSON, pathLabelValues, loggedPath)
		results = append(results, metrics...)
	}

	return results
}

func (pr propertyRule) validate() error {
	if pr.Pointer == "" {
		return errors.New("`Pointer` is mandatory for property rule")
	}
	if pr.Pointer[0] != '/' {
		return errors.New("`Pointer` must begin with '/'")
	}
	if pr.Name == "" {
		return errors.New("`Name` is mandatory for property rule")
	}
	if pr.Type == "" {
		return errors.New("`Type` is mandatory for property rule")
	}

	if _, ok := typeToConverters[pr.Type]; !ok {
		return errors.New("unknown metric type: " + pr.Type)
	}

	return nil
}

func (pr *propertyRule) compile(pathLabelNames []string) error {
	pr.converter = typeToConverters[pr.Type]

	labelNames := getLabelNamesInPath(pr.Pointer)
	allLabelNames := concatenate(pathLabelNames, labelNames)
	pr.desc = prometheus.NewDesc(prometheus.BuildFQName(namespace, "", pr.Name), pr.Help, allLabelNames, nil)

	return nil
}

func (pr propertyRule) matchPointer(parsedJSON *gabs.Container, pathLabelValues []string, loggedPath string) []prometheus.Metric {
	var results []prometheus.Metric

	matchedProperties := pr.matchPointerAux(pr.Pointer, parsedJSON, loggedPath)
	for _, property := range matchedProperties {
		labelValues := concatenate(pathLabelValues, property.indexes)
		m, err := prometheus.NewConstMetric(pr.desc, prometheus.GaugeValue, property.value, labelValues...)
		if err != nil {
			log.Warn("failed to create metric", map[string]interface{}{
				"path":      loggedPath,
				"pointer":   pr.Pointer,
				"name":      pr.Name,
				"value":     property.value,
				log.FnError: err,
			})
			continue
		}

		results = append(results, m)
	}

	return results
}

func (pr propertyRule) matchPointerAux(pointer string, parsedJSON *gabs.Container, loggedPath string) []matchedProperty {
	hasIndexPattern, subPointer, remainder := pr.splitPointer(pointer)
	if !hasIndexPattern {
		v := pr.matchPlainPointer(pointer, parsedJSON)
		if v == nil {
			log.Warn("cannot find pointed value", map[string]interface{}{
				"path":    loggedPath,
				"pointer": pr.Pointer,
			})
			return nil
		}

		value, err := pr.converter(v.Data())
		if err != nil {
			log.Warn("failed to interpret Redfish data as metric", map[string]interface{}{
				"path":      loggedPath,
				"pointer":   pr.Pointer,
				"name":      pr.Name,
				"value":     v.Data(),
				log.FnError: err,
			})
			return nil
		}

		return []matchedProperty{
			{
				value:   value,
				indexes: nil,
			},
		}
	}

	v := pr.matchPlainPointer(subPointer, parsedJSON)
	if v == nil {
		log.Warn("cannot find pointed value", map[string]interface{}{
			"path":    loggedPath,
			"pointer": pr.Pointer,
		})
		return nil
	}

	children, err := v.Children()
	if err != nil {
		log.Warn("index pattern is used, but parent is not array", map[string]interface{}{
			"path":    loggedPath,
			"pointer": pr.Pointer,
		})
		return nil
	}

	var result []matchedProperty
	for i, child := range children {
		ms := pr.matchPointerAux(remainder, child, loggedPath)
		for _, m := range ms {
			m.indexes = append([]string{strconv.Itoa(i)}, m.indexes...)
			result = append(result, m)
		}
	}

	return result
}

func (pr propertyRule) splitPointer(pointer string) (hasIndexPattern bool, subPointer, remainder string) {
	ts := strings.Split(pointer, "/")
	for i, t := range ts {
		if _, ok := getLabelName(t); ok {
			hasIndexPattern = true
			subPointer = strings.Join(ts[0:i], "/")
			if i != len(ts)-1 {
				remainder = "/" + strings.Join(ts[i+1:], "/")
			}
			return
		}
	}
	return false, "", ""
}

func (pr propertyRule) matchPlainPointer(pointer string, parsedJSON *gabs.Container) *gabs.Container {
	p := strings.ReplaceAll(pointer[1:], "/", ".")
	return parsedJSON.Path(p)
}

func concatenate(s, t []string) []string {
	var r []string
	r = append(r, s...)
	return append(r, t...)
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
