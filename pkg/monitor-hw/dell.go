package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"time"

	"github.com/cybozu-go/setup-hw/collector"
	"github.com/cybozu-go/setup-hw/config"
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

func monitorDell(ctx context.Context, ac *config.AddressConfig, uc *config.UserConfig) error {
	if err := initDell(ctx, ac, uc); err != nil {
		return err
	}

	env := well.NewEnvironment(ctx)
	env.Go(func(ctx context.Context) error {
		for {
			values, err := client.Chassis(ctx)
			if err != nil {
				// TODO: log and continue?
				return err
			}
			collector.Metrics.Set("chassis", values)
			select {
			case <-time.After(time.Duration(5 * time.Minute)):
			case <-ctx.Done():
				return nil
			}
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
	mux := http.NewServeMux()
	mux.Handle("/metrics", handler)
	serv := &well.HTTPServer{
		Server: &http.Server{
			Addr:    ":9137",
			Handler: mux,
		},
	}
	err = serv.ListenAndServe()
	if err != nil {
		return err
	}

	env.Stop()
	return env.Wait()
}

func initDell(ctx context.Context, ac *config.AddressConfig, uc *config.UserConfig) error {
	collector.Metrics = collector.NewSafeMetrics()

	// endpoint, err := url.Parse("https://support:" + userConfig.Support.Password.Raw + "@" + addressConfig.IPv4.Address)
	// if err != nil {
	// 	return err
	// }
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
	client = redfish.New(endpoint, transport)

	// TODO: uncomment this
	//return well.CommandContext(ctx, "/usr/libexec/instsvcdrv-helper", "start").Run()
	return nil
}
