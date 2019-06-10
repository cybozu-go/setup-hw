package redfish

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/gabs"
)

type redfishClient struct {
	redfishVersion string
	endpoint       *url.URL
	user           string
	password       string
	httpClient     *http.Client
}

// ClientConfig is a set of configurations for redfishClient.
type ClientConfig struct {
	AddressConfig *config.AddressConfig
	Port          string
	UserConfig    *config.UserConfig
	Rule          *CollectRule
}

// NewRedfishClient create a client for Redfish API
func NewRedfishClient(cc *ClientConfig) (Client, error) {
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

	return &redfishClient{
		endpoint: endpoint,
		user:     "support",
		password: cc.UserConfig.Support.Password.Raw,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   5 * time.Second,
		},
	}, nil
}

func (c *redfishClient) Traverse(ctx context.Context, rule *CollectRule) Collected {
	cl := Collected{data: make(map[string]*gabs.Container), rule: rule}
	c.get(ctx, rule.TraverseRule.Root, cl)
	return cl
}

func (c *redfishClient) GetVersion(ctx context.Context) (string, error) {
	req := c.newRequest(ctx, "/redfish/v1/")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Warn("failed to GET Redfish data", map[string]interface{}{
			"url":       req.URL.String(),
			log.FnError: err,
		})
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Warn("Redfish answered non-OK", map[string]interface{}{
			"url":       req.URL.String(),
			"status":    resp.StatusCode,
			log.FnError: err,
		})
		return "", fmt.Errorf("%d: %s", resp.StatusCode, req.URL.String())
	}

	var result struct {
		RedfishVersion string
	}
	err = json.NewDecoder(resp.Body).Decode(&result)

	return result.RedfishVersion, err
}

func (c *redfishClient) get(ctx context.Context, path string, cl Collected) {
	if cl.rule.TraverseRule.excludeRegexp != nil && cl.rule.TraverseRule.excludeRegexp.MatchString(path) {
		return
	}

	if _, ok := cl.data[path]; ok {
		return
	}

	req := c.newRequest(ctx, path)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Warn("failed to GET Redfish data", map[string]interface{}{
			"url":       req.URL.String(),
			log.FnError: err,
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn("Redfish answered non-OK", map[string]interface{}{
			"url":       req.URL.String(),
			"status":    resp.StatusCode,
			log.FnError: err,
		})
		return
	}

	parsed, err := gabs.ParseJSONBuffer(resp.Body)
	if err != nil {
		log.Warn("failed to parse Redfish data", map[string]interface{}{
			"url":       req.URL.String(),
			log.FnError: err,
		})
		return
	}
	cl.data[path] = parsed

	c.follow(ctx, parsed, cl)
}

func (c *redfishClient) newRequest(ctx context.Context, path string) *http.Request {
	u := *c.endpoint
	u.Path = path

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth(c.user, c.password)
	req.Header.Set("Accept", "application/json")
	req = req.WithContext(ctx)

	return req
}

func (c *redfishClient) follow(ctx context.Context, parsed *gabs.Container, cl Collected) {
	if childrenMap, err := parsed.ChildrenMap(); err == nil {
		for k, v := range childrenMap {
			if k != "@odata.id" {
				c.follow(ctx, v, cl)
			} else if path, ok := v.Data().(string); ok {
				c.get(ctx, path, cl)
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
			c.follow(ctx, v, cl)
		}
		return
	}
}
