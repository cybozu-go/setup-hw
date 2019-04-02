//go:generate statik -f -src ../../../redfish/rules -dest=../../../redfish

package cmd

import (
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/redfish"
	_ "github.com/cybozu-go/setup-hw/redfish/statik" // import for initialization
	"github.com/cybozu-go/well"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rakyll/statik/fs"
)

func startExporter(ac *config.AddressConfig, uc *config.UserConfig, ruleFile string) error {
	statikFS, err := fs.New()
	if err != nil {
		return err
	}
	f, err := statikFS.Open(ruleFile)
	if err != nil {
		return err
	}
	rule, err := ioutil.ReadAll(f)
	if err != nil {
		return err
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
