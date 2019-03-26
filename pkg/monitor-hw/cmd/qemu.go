package cmd

import (
	"context"

	"github.com/cybozu-go/setup-hw/config"
	"github.com/prometheus/client_golang/prometheus"
)

func monitorQEMU(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func initExporterQEMU(ac *config.AddressConfig, uc *config.UserConfig) (prometheus.Collector, error) {
	// TODO: return valid Collector
	return nil, nil
}
