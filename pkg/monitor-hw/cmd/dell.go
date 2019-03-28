package cmd

import (
	"context"
	"time"

	"github.com/cybozu-go/setup-hw/collector"
	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/well"
	"github.com/prometheus/client_golang/prometheus"
)

func monitorDell(ctx context.Context) error {
	if err := initDell(ctx); err != nil {
		return err
	}

	env := well.NewEnvironment(ctx)
	env.Go(func(ctx context.Context) error {
		for {
			select {
			case <-time.After(time.Duration(opts.resetInterval) * time.Second):
			case <-ctx.Done():
				return nil
			}
			// TODO: reset iDRAC
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
