package cmd

import (
	"context"
	"net/http"
	"time"

	"github.com/cybozu-go/well"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func startExporter(collector prometheus.Collector) error {
	well.Go(func(ctx context.Context) error {
		for {
			// TODO: parallelize
			for _, c := range opts.collectors {
				err := c.Update(ctx)
				if err != nil {
					// TODO: log and continue?
					return err
				}
			}
			select {
			case <-time.After(time.Duration(opts.interval) * time.Second):
			case <-ctx.Done():
				return nil
			}
		}
		return nil
	})

	err := prometheus.Register(collector)
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
