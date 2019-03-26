package main

import (
	"context"

	"github.com/cybozu-go/setup-hw/config"
)

func monitorQEMU(ctx context.Context, ac *config.AddressConfig, uc *config.UserConfig) error {
	<-ctx.Done()
	return nil
}
