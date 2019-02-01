package main

import (
	"context"
	"errors"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/lib"
	"github.com/cybozu-go/well"
)

func main() {
	well.LogConfig{}.Apply()

	vendor, err := lib.DetectVendor()
	if err != nil {
		log.ErrorExit(err)
	}

	var monitor func(ctx context.Context) error
	switch vendor {
	case lib.QEMU:
		monitor = monitorQEMU
	case lib.Dell:
		monitor = monitorDell
	default:
		log.ErrorExit(errors.New("unsupported vendor hardware"))
	}

	well.Go(monitor)
	well.Stop()
	err = well.Wait()
	if err != nil && !well.IsSignaled(err) {
		log.ErrorExit(err)
	}
}
