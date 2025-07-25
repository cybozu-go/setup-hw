package redfish

import (
	"errors"
	"fmt"
	"math"
)

type converter func(interface{}) (float64, error)

var typeToConverters = map[string]converter{
	"number": numberConverter,
	"health": healthConverter,
	"state":  stateConverter,
	"bool":   boolConverter,
}

func numberConverter(data interface{}) (float64, error) {
	// Redfish sometimes returns null for numeric values, so we handle that case.
	if data == nil {
		return math.NaN(), nil
	}
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

func boolConverter(data interface{}) (float64, error) {
	if data == nil {
		return -1, nil
	}
	health, ok := data.(bool)
	if !ok {
		return -1, errors.New("health value was not boolean")
	}
	switch health {
	case false:
		return 0, nil
	case true:
		return 1, nil
	}
	return -1, fmt.Errorf("unknown health value: %t", health)
}
