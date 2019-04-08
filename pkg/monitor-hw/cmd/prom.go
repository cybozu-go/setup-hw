package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/redfish"
	"github.com/cybozu-go/well"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type logger struct{}

func (l logger) Println(v ...interface{}) {
	log.Error(fmt.Sprint(v...), nil)
}

func startExporter(ac *config.AddressConfig, uc *config.UserConfig, ruleFile string) error {
	rule, ok := redfish.Rules[ruleFile]
	if !ok {
		return errors.New("unknown rule file: " + ruleFile)
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
			collector.Update(ctx)
			select {
			case <-time.After(time.Duration(opts.interval) * time.Second):
			case <-ctx.Done():
				return nil
			}
		}
	})

	registry := prometheus.NewRegistry()
	err = registry.Register(collector)
	if err != nil {
		return err
	}

	handler := promhttp.HandlerFor(registry,
		promhttp.HandlerOpts{
			ErrorLog:      logger{},
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
