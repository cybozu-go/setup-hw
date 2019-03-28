package redfish

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Jeffail/gabs"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/config"
)

// Redfish contains the Endpoint and a Client
type Redfish struct {
	Endpoint *url.URL
	Client   *http.Client
}

// {
// 	Path: "/redfish/v1/system/{sid}",
// 	Pointer: "/processor[{pid}]/status",
// 	Name: "processor_status",
// 	Type: StatusConverter
// }

// Value contains a metrics name, value and labels
type Value struct {
	Name   string
	Value  float64
	Labels map[string]string
}

// Status represents Status type of Redfish
type Status struct {
	Health       string `json:"Health"`
	HealthRollup string `json:"HealthRollup"`
	State        string `json:"State"`
}

var rfclient *Redfish

// New returns a new *Redfish
func New(ac *config.AddressConfig, uc *config.UserConfig) (*Redfish, error) {
	endpoint, err := url.Parse("https://" + ac.IPv4.Address)
	if err != nil {
		return nil, err
	}
	endpoint.User = url.UserPassword("support", uc.Support.Password.Raw)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return &Redfish{
		Endpoint: endpoint,
		Client: &http.Client{
			Transport: transport,
		},
	}, nil
}

func (r *Redfish) get(ctx context.Context, path string, cmap ContainerMap) ContainerMap {
	u, err := r.Endpoint.Parse(path)
	if err != nil {
		log.Warn("failed to parse Redfish path", map[string]interface{}{
			"path":      path,
			log.FnError: err,
		})
		return cmap
	}

	epath := u.EscapedPath()
	if _, ok := cmap[epath]; ok {
		return cmap
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Warn("failed to create GET request", map[string]interface{}{
			"url":       u.String(),
			log.FnError: err,
		})
		return cmap
	}
	req = req.WithContext(ctx)

	resp, err := r.Client.Do(req)
	if err != nil {
		log.Warn("failed to GET Redfish data", map[string]interface{}{
			"url":       u.String(),
			log.FnError: err,
		})
		return cmap
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn("Redfish answered non-OK", map[string]interface{}{
			"url":       u.String,
			"status":    resp.StatusCode,
			log.FnError: err,
		})
		return cmap
	}

	parsed, err := gabs.ParseJSON(ioutil.ReadAll(resp.Body))
	if err != nil {
		log.Warn("failed to parse Redfish data", map[string]interface{}{
			"url":       u.String(),
			log.FnError: err,
		})
		return cmap
	}
	cmap[epath] = parsed

	return r.follow(ctx, parsed, cmap)
}

func (r *Redfish) follow(ctx context.Context, parsed *gabs.Container, cmap ContainerMap) ContainerMap {
	childrenMap, ok := parsed.ChildrenMap()
	if ok {
		for k, v := range childrenMap {
			if k == "@odata.id" {
				path, ok := v.Data().(string)
				if !ok {
					log.Warn("value of @odata.id is not string", map[string]interface{}{
						"value": v.String(),
					})
					continue
				}
				cmap = r.get(ctx, path, cmap)
			} else {
				cmap = r.follow(ctx, v, cmap)
			}
		}
		return cmap
	}

	childrenSlice, ok := parsed.Children()
	if ok {
		for _, v := range childrenSlice {
			cmap = r.follow(ctx, v, cmap)
		}
		return cmap
	}

	return cmap
}

func (r *Redfish) Traverse(ctx context.Context, rootPath str) (ContainerMap, error) {
	return r.get(ctx, rootPath, make(ContainerMap))
}
