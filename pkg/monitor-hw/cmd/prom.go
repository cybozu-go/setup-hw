package cmd

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/redfish"
	"github.com/cybozu-go/well"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func startExporter(ac *config.AddressConfig, uc *config.UserConfig, ruleFile string) error {
	var rule io.Reader
	if ruleFile != "" {
		f, err := os.Open(ruleFile) // TODO: use statik
		if err != nil {
			return err
		}
		defer f.Close()
		rule = f
	} else {
		rule = bytes.NewBuffer(nil)
	}

	cc := &redfish.CollectorConfig{
		AddressConfig: ac,
		UserConfig:    uc,
		Rule:          rule,
	}

	collector, err := redfish.NewCollector(cc)
	if err != nil {
		return err
	}

	well.Go(func(ctx context.Context) error {
		for {
			collector.Update(ctx, opts.redfishRoot)
			select {
			case <-time.After(time.Duration(opts.interval) * time.Second):
			case <-ctx.Done():
				return nil
			}
		}
		return nil
	})

	err = prometheus.Register(collector)
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
			Addr:    opts.listenAddress,
			Handler: mux,
		},
	}
	return serv.ListenAndServe()
}
