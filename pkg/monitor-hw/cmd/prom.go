package cmd

import (
	"context"
	"net/http"
	"time"

	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/redfish"
	"github.com/cybozu-go/well"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func startExporter(ac *config.AddressConfig, uc *config.UserConfig, ruleFile string) error {
	rfclient, err := redfish.NewRedfish(ac, uc)
	if err != nil {
		return err
	}

	well.Go(func(ctx context.Context) error {
		for {
			rfclient.Update(ctx, opts.redfishRoot)

			select {
			case <-time.After(time.Duration(opts.interval) * time.Second):
			case <-ctx.Done():
				return nil
			}
		}
		return nil
	})

	collector, err := redfish.NewRedfishCollector(ruleFile)
	if err != nil {
		return err
	}
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
