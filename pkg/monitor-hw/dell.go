package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/cybozu-go/setup-hw/redfish"
	"github.com/cybozu-go/well"
)

var client *redfish.Redfish

func monitorDell(ctx context.Context) error {
	if err := initDell(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-time.After(time.Duration(5 * time.Minute)):
		case <-ctx.Done():
			return nil
		}
		values, err := client.Chassis(ctx)
		if err != nil {
			return err
		}
		fmt.Println(values)
	}

	return nil
}

func initDell(ctx context.Context) error {
	data, err := ioutil.ReadFile("/etc/neco/bmc-address.json")
	if err != nil {
		return err
	}
	addressConfig := AddressConfig{}
	if err := json.Unmarshal(data, &addressConfig); err != nil {
		return err
	}

	data, err = ioutil.ReadFile("/etc/neco/bmc-user.json")
	if err != nil {
		return err
	}
	userConfig := UserConfig{}
	if err := json.Unmarshal(data, &userConfig); err != nil {
		return err
	}

	endpoint, err := url.Parse("https://support:" + userConfig.Support.Password.Raw + "@" + addressConfig.IPv4.Address)
	if err != nil {
		return err
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client = redfish.New(endpoint, transport)

	return well.CommandContext(ctx, "/usr/libexec/instsvcdrv-helper", "start").Run()
}
