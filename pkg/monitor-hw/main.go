package main

import (
	"context"
	"errors"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/lib"
	"github.com/cybozu-go/well"
)

func main() {
	well.LogConfig{}.Apply()

	ac, uc, err := config.LoadConfig()
	if err != nil {
		log.ErrorExit(err)
	}

	vendor, err := lib.DetectVendor()
	if err != nil {
		log.ErrorExit(err)
	}

	var monitor func(context.Context, *config.AddressConfig, *config.UserConfig) error
	switch vendor {
	case lib.QEMU:
		monitor = monitorQEMU
	case lib.Dell:
		monitor = monitorDell
	default:
		log.ErrorExit(errors.New("unsupported vendor hardware"))
	}

	well.Go(func(ctx context.Context) error {
		return monitor(ctx, ac, uc)
	})
	well.Stop()
	err = well.Wait()
	if err != nil && !well.IsSignaled(err) {
		log.ErrorExit(err)
	}
}
