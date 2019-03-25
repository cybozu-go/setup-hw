package redfish

// Options allow to set options for the Redfish package
type Options struct {
	RedfishEndpoint string
}

// Redfish contains the Options and a Client to mock outputs during development
type Redfish struct {
	Endpoint string
}

// Value contains a metrics name, value and labels
type Value struct {
	Name   string
	Value  string
	Labels map[string]string
}

// New returns a new *Redfish
func New(endpoint string) *Redfish {
	return &Redfish{
		Endpoint: endpoint,
	}
}
