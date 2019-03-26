package collector

import (
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/redfish"
)

var rfclient *redfish.Redfish

// InitRedfishCollectors initializes
func InitRedfishCollectors(ac *config.AddressConfig, uc *config.UserConfig) error {
	endpoint, err := url.Parse("https://" + ac.IPv4.Address)
	if err != nil {
		return err
	}
	endpoint.User = url.UserPassword("support", uc.Support.Password.Raw)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	rfclient = redfish.New(endpoint, transport)

	return nil
}
