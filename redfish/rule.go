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

type converter func(interface{}) (float64, error)

func (cr convertRule) Validate() error {
	if cr.Path == "" {
		return errors.New("Path is mandatory for convert rule")
	}

	for _, propertyRule := range cr.Rules {
		if err := propertyRule.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (cpr convertPropertyRule) Validate() error {
	if cpr.Pointer == "" {
		return errors.New("Pointer is mandatory for convert property rule")
	}
	if cpr.Name == "" {
		return errors.New("Name is mandatory for convert property rule")
	}
	if cpr.Converter == nil {
		return errors.New("Converter is mandatory for convert property rule")
	}

	return nil
}

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
	"value":  valueConverter,
	"health": healthConverter,
	"state":  stateConverter,
}

func valueConverter(data interface{}) (float64, error) {
	value, ok := data.(float64)
	if !ok {
		return 0, errors.New("value was not float64")
	}
	return value, nil
}

func healthConverter(data interface{}) (float64, error) {
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
