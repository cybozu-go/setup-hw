package redfish

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/gabs"
)

type client struct {
	endpoint   *url.URL
	httpClient *http.Client
	cache      *cache
}

func newClient(cc *CollectorConfig, cache *cache) (*client, error) {
	endpoint, err := url.Parse("https://" + cc.AddressConfig.IPv4.Address)
	if err != nil {
		return nil, err
	}
	endpoint.User = url.UserPassword("support", cc.UserConfig.Support.Password.Raw)
	if cc.Port != "" {
		endpoint.Host = endpoint.Host + ":" + cc.Port
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return &client{
		endpoint: endpoint,
		httpClient: &http.Client{
			Transport: transport,
		},
		cache: cache,
	}, nil
}

func (c *client) get(ctx context.Context, path string, dataMap dataMap) dataMap {
	u, err := c.endpoint.Parse(path)
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

	resp, err := c.httpClient.Do(req)
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
			"url":       u.String(),
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

	return c.follow(ctx, parsed, dataMap)
}

func (c *client) follow(ctx context.Context, parsed *gabs.Container, dataMap dataMap) dataMap {
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
				dataMap = c.get(ctx, path, dataMap)
			} else {
				dataMap = c.follow(ctx, v, dataMap)
			}
		}
		return dataMap
	}

	if childrenSlice, err := parsed.Children(); err == nil {
		for _, v := range childrenSlice {
			dataMap = c.follow(ctx, v, dataMap)
		}
		return dataMap
	}

	return dataMap
}

func (c *client) update(ctx context.Context, rootPath string) {
	dataMap := c.get(ctx, rootPath, make(dataMap))
	c.cache.set(dataMap)
}
