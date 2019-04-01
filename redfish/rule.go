package redfish

import "errors"

type convertRule struct {
	Path  string                `yaml:"Path"`
	Rules []convertPropertyRule `yaml:"Rules"`
}

type convertPropertyRule struct {
	Pointer     string    `yaml:"Pointer"`
	Name        string    `yaml:"Name"`
	Description string    `yaml:"Description"`
	Converter   converter `yaml:"Type"`
}

type converter func(interface{}) float64

func (c *converter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var typeName string
	err := unmarshal(&typeName)
	if err != nil {
		return err
	}

	converter, ok := typeToConverters[typeName]
	if !ok {
		return errors.New("unknown metrics type: " + typeName)
	}

	*c = converter
	return nil
}

var typeToConverters = map[string]converter{
	"health": healthConverter,
}

func healthConverter(data interface{}) float64 {
	health, ok := data.(string)
	if !ok {
		return -1
	}
	switch health {
	case "OK":
		return 0
	case "Warning":
		return 1
	case "Critical":
		return 2
	}
	return -1
}
