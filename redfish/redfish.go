package redfish

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/gabs"
)

type redfishClient struct {
	endpoint   *url.URL
	user       string
	password   string
	httpClient *http.Client
	noEscape   bool
	token      string
	sessionId  string
}

// ClientConfig is a set of configurations for redfishClient.
type ClientConfig struct {
	AddressConfig *config.AddressConfig
	Port          string
	UserConfig    *config.UserConfig
	Rule          *CollectRule
	NoEscape      bool
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
		noEscape: cc.NoEscape,
	}, nil
}

func (c *redfishClient) Traverse(ctx context.Context, rule *CollectRule) Collected {
	cl := Collected{data: make(map[string]*gabs.Container), rule: rule}
	c.get(ctx, rule.TraverseRule.Root, cl)
	return cl
}

func (c *redfishClient) GetVersion(ctx context.Context) (string, error) {
	req, err := c.newRequest(ctx, "/redfish/v1/")
	if err != nil {
		panic(err)
	}

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
	if !cl.rule.TraverseRule.NeedTraverse(path) {
		return
	}

	if _, ok := cl.data[path]; ok {
		return
	}

	req, err := c.newRequest(ctx, path)
	if err != nil {
		log.Warn("failed to create request", map[string]interface{}{
			"path":      path,
			log.FnError: err,
		})
		return
	}

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

func (c *redfishClient) newRequest(ctx context.Context, path string) (*http.Request, error) {
	p := path
	if !c.noEscape {
		p = url.PathEscape(p)
	}
	u, err := c.endpoint.Parse(p)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-auth-token", c.token)
	req.Header.Set("Accept", "application/json")
	req = req.WithContext(ctx)
	return req, nil
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

type sessionLoginRequest struct {
	Username string `json:"UserName"`
	Password string `json:"Password"`
}

type sessionLoginResponse struct {
	Id string `json:"Id"`
}

func (c *redfishClient) Login(ctx context.Context) error {
	if c.checkSession(ctx) == nil {
		return nil
	}
	bAuth := sessionLoginRequest{
		Username: c.user,
		Password: c.password,
	}
	jsonBody, err := json.Marshal(bAuth)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %s", err)
	}

	u := c.endpoint.JoinPath("/redfish/v1/SessionService/Sessions")
	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("error making NewRequest: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error accessing rest service: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("error response from rest service: %d", resp.StatusCode)
	}

	byteJSON, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading http body: %s", err)
	}

	var ses sessionLoginResponse
	err = json.Unmarshal(byteJSON, &ses)
	if err != nil {
		return fmt.Errorf("error when unmarshaling iDRAC response: %s", err)
	}
	c.sessionId = ses.Id
	c.token = resp.Header.Get("X-auth-token")
	return nil
}

func (c *redfishClient) checkSession(ctx context.Context) error {
	u := c.endpoint.JoinPath("/redfish/v1/SessionService/Sessions", c.sessionId)
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return fmt.Errorf("error making NewRequest: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-auth-token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error accessing Redfish service: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error response from Redfish service: %d", resp.StatusCode)
	}
	return nil
}
