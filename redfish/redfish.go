package redfish

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

// Options allow to set options for the Redfish package
type Options struct {
	RedfishEndpoint string
}

// Redfish contains the Options and a Client to mock outputs during development
type Redfish struct {
	Endpoint *url.URL
	Client   *http.Client
}

// Value contains a metrics name, value and labels
type Value struct {
	Name   string
	Value  string
	Labels map[string]string
}

// Status represents Status type of Redfish
type Status struct {
	Health       string `json:"Health"`
	HealthRollup string `json:"HealthRollup"`
	State        string `json:"State"`
}

func convertHealth(health string) string {
	switch health {
	case "OK":
		return "0"
	case "Warning":
		return "1"
	case "Critical":
		return "2"
	}
	return "-1"
}

func convertState(state string) string {
	switch state {
	case "Enabled":
		return "0"
	case "Disabled":
		return "1"
	}
	return "-1"
}

// New returns a new *Redfish
func New(endpoint *url.URL, transport *http.Transport) *Redfish {
	return &Redfish{
		Endpoint: endpoint, // https://user:pass@1.2.3.4
		Client: &http.Client{
			Transport: transport,
		},
	}
}

func (r *Redfish) get(ctx context.Context, path string) ([]byte, error) {
	u, err := r.Endpoint.Parse(path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Status not OK: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

// Chassis returns metrics values of chassis
func (r *Redfish) Chassis(ctx context.Context) ([]Value, error) {
	values := make([]Value, 0)

	data, err := r.get(ctx, "/redfish/v1/Chassis")
	if err != nil {
		return nil, err
	}

	chassis := new(struct {
		Members []struct {
			ODataID string `json:"@odata.id"`
		} `json:"Members"`
	})
	err = json.Unmarshal(data, chassis)
	if err != nil {
		return nil, err
	}

	for _, m := range chassis.Members {
		chassisID := path.Base(m.ODataID)

		data, err := r.get(ctx, m.ODataID)
		c := new(struct {
			Status Status `json:"Status"`
		})
		err = json.Unmarshal(data, c)
		if err != nil {
			return nil, err
		}
		values = append(values,
			Value{
				Name:   "chassis_status_health",
				Value:  convertHealth(c.Status.Health),
				Labels: map[string]string{"Chassis": chassisID},
			},
			Value{
				Name:   "chassis_status_healthrollup",
				Value:  convertHealth(c.Status.HealthRollup),
				Labels: map[string]string{"Chassis": chassisID},
			},
			Value{
				Name:   "chassis_status_state",
				Value:  convertState(c.Status.State),
				Labels: map[string]string{"Chassis": chassisID},
			},
		)
	}

	return values, nil
}
