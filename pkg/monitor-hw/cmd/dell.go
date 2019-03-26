package cmd

import (
	"context"
	"time"

	"github.com/cybozu-go/setup-hw/collector"
	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/well"
	"github.com/prometheus/client_golang/prometheus"
)

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

	env := well.NewEnvironment(ctx)
	env.Go(func(ctx context.Context) error {
		for {
			// TODO: reset iDRAC
			select {
			case <-time.After(time.Duration(opts.resetInterval) * time.Second):
			case <-ctx.Done():
				return nil
			}
		}
		return nil
	})

	env.Stop()
	return env.Wait()
}

func initDell(ctx context.Context) error {
	// TODO: uncomment this
	//return well.CommandContext(ctx, "/usr/libexec/instsvcdrv-helper", "start").Run()
	return nil
}

func initExporterDell(ac *config.AddressConfig, uc *config.UserConfig) (prometheus.Collector, error) {
	err := collector.InitRedfishCollectors(ac, uc)
	if err != nil {
		return nil, err
	}

	return RedfishCollector{collectors: opts.collectors}, nil
}
