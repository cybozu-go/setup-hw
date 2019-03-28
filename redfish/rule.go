package redfish

import "errors"

type ConvertRule struct {
	Path  string                `yaml:"Path"`
	Rules []ConvertPropertyRule `yaml:"Rules"`
}

type ConvertPropertyRule struct {
	Pointer     string    `yaml:"Pointer"`
	Name        string    `yaml:"Name"`
	Description string    `yaml:"Description"`
	Converter   Converter `yaml:"Type"`
}

type Converter func(interface{}) float64

func (c *Converter) UnmarshalYAML(unmarshal func(interface{}) error) error {
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

var typeToConverters = map[string]Converter{
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
