package redfish

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Redfish contains the Endpoint and a Client
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

func (s Status) healthValue() string {
	switch s.Health {
	case "OK":
		return "0"
	case "Warning":
		return "1"
	case "Critical":
		return "2"
	}
	return "-1"
}

func (s Status) healthRollupValue() string {
	switch s.HealthRollup {
	case "OK":
		return "0"
	case "Warning":
		return "1"
	case "Critical":
		return "2"
	}
	return "-1"
}

func (s Status) stateValue() string {
	switch s.State {
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
		Endpoint: endpoint,
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
