package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/cybozu-go/setup-hw/collector"
	"github.com/cybozu-go/setup-hw/redfish"
	"github.com/cybozu-go/well"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var client *redfish.Redfish

type RedfishCollector struct {
	collectors map[string]collector.Collector
}

func (c RedfishCollector) Describe(ch chan<- *prometheus.Desc) {
}

func (c RedfishCollector) Collect(ch chan<- prometheus.Metric) {
	for _, c := range c.collectors {
		c.Collect(ch)
	}
}

func monitorDell(ctx context.Context) error {
	if err := initDell(ctx); err != nil {
		return err
	}

	well.Go(func(ctx context.Context) error {
		for {
			select {
			case <-time.After(time.Duration(5 * time.Minute)):
			case <-ctx.Done():
				return nil
			}
			values, err := client.Chassis(ctx)
			if err != nil {
				// TODO: log and continue?
				return err
			}
			collector.Metrics.Set("chassis", values)
		}
		return nil
	})

	collectors := make(map[string]collector.Collector)
	collectors["chassis"] = collector.NewChassisCollector()
	err := prometheus.Register(RedfishCollector{collectors: collectors})
	if err != nil {
		return err
	}

	handler := promhttp.HandlerFor(prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			ErrorLog:      nil, // TODO
			ErrorHandling: promhttp.ContinueOnError,
		})
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	})

	return http.ListenAndServe(":9137", nil)
}

func initDell(ctx context.Context) error {
	collector.Metrics = collector.NewSafeMetrics()

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

	//return well.CommandContext(ctx, "/usr/libexec/instsvcdrv-helper", "start").Run()
	return nil
}
