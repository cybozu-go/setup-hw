package redfish

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/gabs"
)

// Redfish contains the Endpoint and a Client
type Redfish struct {
	Endpoint *url.URL
	Client   *http.Client
	cache    *Cache
}

// New returns a new *Redfish
func NewRedfish(ac *config.AddressConfig, uc *config.UserConfig, cache *Cache) (*Redfish, error) {
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
		cache: cache,
	}, nil
}

func (r *Redfish) get(ctx context.Context, path string, dataMap RedfishDataMap) RedfishDataMap {
	u, err := r.Endpoint.Parse(path)
	if err != nil {
		log.Warn("failed to parse Redfish path", map[string]interface{}{
			"path":      path,
			log.FnError: err,
		})
		return dataMap
	}

	epath := u.EscapedPath()
	if _, ok := dataMap[epath]; ok {
		return dataMap
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Warn("failed to create GET request", map[string]interface{}{
			"url":       u.String(),
			log.FnError: err,
		})
		return dataMap
	}
	req = req.WithContext(ctx)

	resp, err := r.Client.Do(req)
	if err != nil {
		log.Warn("failed to GET Redfish data", map[string]interface{}{
			"url":       u.String(),
			log.FnError: err,
		})
		return dataMap
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn("Redfish answered non-OK", map[string]interface{}{
			"url":       u.String,
			"status":    resp.StatusCode,
			log.FnError: err,
		})
		return dataMap
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warn("failed to read Redfish data", map[string]interface{}{
			"url":       u.String(),
			log.FnError: err,
		})
		return dataMap
	}

	parsed, err := gabs.ParseJSON(body)
	if err != nil {
		log.Warn("failed to parse Redfish data", map[string]interface{}{
			"url":       u.String(),
			log.FnError: err,
		})
		return dataMap
	}
	dataMap[epath] = parsed

	return r.follow(ctx, parsed, dataMap)
}

func (r *Redfish) follow(ctx context.Context, parsed *gabs.Container, dataMap RedfishDataMap) RedfishDataMap {
	if childrenMap, err := parsed.ChildrenMap(); err == nil {
		for k, v := range childrenMap {
			if k == "@odata.id" {
				path, ok := v.Data().(string)
				if !ok {
					log.Warn("value of @odata.id is not string", map[string]interface{}{
						"value": v.String(),
					})
					continue
				}
				dataMap = r.get(ctx, path, dataMap)
			} else {
				dataMap = r.follow(ctx, v, dataMap)
			}
		}
		return dataMap
	}

	if childrenSlice, err := parsed.Children(); err == nil {
		for _, v := range childrenSlice {
			dataMap = r.follow(ctx, v, dataMap)
		}
		return dataMap
	}

	return dataMap
}

func (r *Redfish) Update(ctx context.Context, rootPath string) {
	dataMap := r.get(ctx, rootPath, make(RedfishDataMap))
	r.cache.Set(dataMap)
}
