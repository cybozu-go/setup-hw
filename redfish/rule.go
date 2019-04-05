package redfish

import (
	"errors"
	"regexp"
	"strings"

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

type converter func(interface{}) (float64, error)

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
		return errors.New("Root is mandatory for traverse rule")
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
		return errors.New("Path is mandatory for metric rule")
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

func (pr propertyRule) validate() error {
	if pr.Pointer == "" {
		return errors.New("Pointer is mandatory for property rule")
	}
	if pr.Name == "" {
		return errors.New("Name is mandatory for property rule")
	}
	if pr.Type == "" {
		return errors.New("Type is mandatory for property rule")
	}

	if _, ok := typeToConverters[pr.Type]; !ok {
		return errors.New("unknown metric type: " + pr.Type)
	}

	return nil
}

func (pr *propertyRule) compile(pathLabelNames []string) error {
	pr.converter = typeToConverters[pr.Type]

	labelNames := getLabelNamesInPath(pr.Pointer)

	var allLabelNames []string
	allLabelNames = append(allLabelNames, pathLabelNames...)
	allLabelNames = append(allLabelNames, labelNames...)
	pr.desc = prometheus.NewDesc(prometheus.BuildFQName(namespace, "", pr.Name), pr.Help, allLabelNames, nil)

	return nil
}

var typeToConverters = map[string]converter{
	"number": numberConverter,
	"health": healthConverter,
	"state":  stateConverter,
}

func numberConverter(data interface{}) (float64, error) {
	value, ok := data.(float64)
	if !ok {
		return 0, errors.New("value was not float64")
	}
	return value, nil
}

func healthConverter(data interface{}) (float64, error) {
	if data == nil {
		return -1, nil
	}
	health, ok := data.(string)
	if !ok {
		return -1, errors.New("health value was not string")
	}
	switch health {
	case "OK":
		return 0, nil
	case "Warning":
		return 1, nil
	case "Critical":
		return 2, nil
	}
	return -1, errors.New("unknown health value: " + health)
}

func stateConverter(data interface{}) (float64, error) {
	state, ok := data.(string)
	if !ok {
		return -1, errors.New("state value was not string")
	}
	switch state {
	case "Enabled":
		return 0, nil
	case "Disabled":
		return 1, nil
	case "Absent":
		return 2, nil
	case "Deferring":
		return 3, nil
	case "InTest":
		return 4, nil
	case "Quiesced":
		return 5, nil
	case "StandbyOffline":
		return 6, nil
	case "StandbySpare":
		return 7, nil
	case "Starting":
		return 8, nil
	case "UnavailableOffline":
		return 9, nil
	case "Updating":
		return 10, nil
	}
	return -1, errors.New("unknown state value: " + state)
}
