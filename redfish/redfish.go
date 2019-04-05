package redfish

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/gabs"
)

type client struct {
	endpoint   *url.URL
	user       string
	password   string
	httpClient *http.Client
}

func newClient(cc *CollectorConfig) (*client, error) {
	endpoint, err := url.Parse("https://" + cc.AddressConfig.IPv4.Address)
	if err != nil {
		return nil, err
	}

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
		user:     "support",
		password: cc.UserConfig.Support.Password.Raw,
		httpClient: &http.Client{
			Transport: transport,
		},
	}, nil
}

func (c *client) traverse(ctx context.Context, rootPath string) dataMap {
	dataMap := make(dataMap)
	c.get(ctx, rootPath, dataMap)
	return dataMap
}

func (c *client) get(ctx context.Context, path string, dataMap dataMap) {
	u, err := c.endpoint.Parse(path)
	if err != nil {
		log.Warn("failed to parse Redfish path", map[string]interface{}{
			"path":      path,
			log.FnError: err,
		})
		return
	}

	epath := u.EscapedPath()
	if _, ok := dataMap[epath]; ok {
		return
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Warn("failed to create GET request", map[string]interface{}{
			"url":       u.String(),
			log.FnError: err,
		})
		return
	}
	req.SetBasicAuth(c.user, c.password)
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Warn("failed to GET Redfish data", map[string]interface{}{
			"url":       u.String(),
			log.FnError: err,
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn("Redfish answered non-OK", map[string]interface{}{
			"url":       u.String(),
			"status":    resp.StatusCode,
			log.FnError: err,
		})
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warn("failed to read Redfish data", map[string]interface{}{
			"url":       u.String(),
			log.FnError: err,
		})
		return
	}

	parsed, err := gabs.ParseJSON(body)
	if err != nil {
		log.Warn("failed to parse Redfish data", map[string]interface{}{
			"url":       u.String(),
			log.FnError: err,
		})
		return
	}
	dataMap[epath] = parsed

	c.follow(ctx, parsed, dataMap)
}

func (c *client) follow(ctx context.Context, parsed *gabs.Container, dataMap dataMap) {
	if childrenMap, err := parsed.ChildrenMap(); err == nil {
		for k, v := range childrenMap {
			if k != "@odata.id" {
				c.follow(ctx, v, dataMap)
			} else if path, ok := v.Data().(string); ok {
				c.get(ctx, path, dataMap)
			} else {
				log.Warn("value of @odata.id is not string", map[string]interface{}{
					"typ":   reflect.TypeOf(v.Data()),
					"value": v.Data(),
				})
			}
		}
		return
	}

	if childrenSlice, err := parsed.Children(); err == nil {
		for _, v := range childrenSlice {
			c.follow(ctx, v, dataMap)
		}
		return
	}
}
