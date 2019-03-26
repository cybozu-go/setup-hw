package redfish

import (
	"context"
	"encoding/json"
	"path"
)

// ChassisSet represents a set of chassis.
type ChassisSet struct {
	Members []struct {
		ODataID string `json:"@odata.id"`
	} `json:"Members"`
}

// Chassis represents a chassis.
type Chassis struct {
	Status Status `json:"Status"`
}

// Chassis returns metrics values of chassis.
func (r *Redfish) Chassis(ctx context.Context) ([]Value, error) {
	values := make([]Value, 0)

	data, err := r.get(ctx, "/redfish/v1/Chassis")
	if err != nil {
		return nil, err
	}

	chassisSet := new(ChassisSet)
	err = json.Unmarshal(data, chassisSet)
	if err != nil {
		return nil, err
	}

	for _, m := range chassisSet.Members {
		chassisID := path.Base(m.ODataID)

		data, err := r.get(ctx, m.ODataID)
		chassis := new(Chassis)
		err = json.Unmarshal(data, chassis)
		if err != nil {
			return nil, err
		}
		labels := map[string]string{"Chassis": chassisID}
		values = append(values,
			Value{
				Name:   "chassis_status_health",
				Value:  chassis.Status.healthValue(),
				Labels: labels,
			},
			Value{
				Name:   "chassis_status_healthrollup",
				Value:  chassis.Status.healthRollupValue(),
				Labels: labels,
			},
			Value{
				Name:   "chassis_status_state",
				Value:  chassis.Status.stateValue(),
				Labels: labels,
			},
		)
	}

	return values, nil
}
