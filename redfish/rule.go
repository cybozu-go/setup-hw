package redfish

type ConvertRule struct {
	Path  string
	Rules []ConvertPropertyRule
}

type ConvertPropertyRule struct {
	Pointer string
	Name    string
	Type    Converter
}

var types = map[string]Converter{
	"health": healthConverter,
}

type Converter func(interface{}) (float64, map[string]string)

func healthConverter(data interface{}) (float64, map[string]string) {
	health, ok := data.(string)
	if !ok {
		return -1, nil
	}
	switch health {
	case "OK":
		return 0, nil
	case "Warning":
		return 1, nil
	case "Critical":
		return 2, nil
	}
	return -1, nil
}
