package lib

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/cybozu-go/setup-hw/config"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// Vendor represents server hardware vendor.
type Vendor int

// Vendors
const (
	Unknown Vendor = iota
	QEMU
	Dell
)

// DetectVendor detects hardware vendor.
func DetectVendor() (Vendor, error) {
	data, err := ioutil.ReadFile("/sys/devices/virtual/dmi/id/sys_vendor")
	if err != nil {
		return Unknown, err
	}

	vendor := strings.TrimSpace(string(data))
	if vendor == "QEMU" {
		return QEMU, nil
	}

	if strings.HasPrefix(vendor, "Dell") {
		return Dell, nil
	}

	return Unknown, errors.New("unknown vendor: " + vendor)
}

// DetectRedfishVersion fetches Redfish version from web API
func DetectRedfishVersion(ac *config.AddressConfig, uc *config.UserConfig) (string, error) {
	endpoint, err := url.Parse("https://" + ac.IPv4.Address)
	if err != nil {
		return "", err
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	endpoint.Path = path.Join(endpoint.Path, "/redfish/v1/")
	req, err := http.NewRequest("GET", endpoint.String(), nil)
	req.SetBasicAuth("support", uc.Support.Password.Raw)
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var version struct {
		RedfishVersion string `json:"RedfishVersion"`
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(data, &version)
	if err != nil {
		return "", err
	}
	return version.RedfishVersion, nil
}
