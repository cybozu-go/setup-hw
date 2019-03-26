package main

import (
	"errors"
	"os"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/lib"
	"github.com/cybozu-go/well"
)

const (
	// ExitReboot is the status code to tell the caller to reboot.
	ExitReboot = 10
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

	var setup func(*config.AddressConfig, *config.UserConfig) (bool, error)
	switch vendor {
	case lib.QEMU:
		setup = setupQEMU
	case lib.Dell:
		setup = setupDell
	default:
		log.ErrorExit(errors.New("unsupported vendor hardware"))
	}

	reboot, err := setup(ac, uc)
	if err != nil {
		log.ErrorExit(err)
	}

	if reboot {
		log.Warn("reboot the server now", nil)
		os.Exit(ExitReboot)
	}
}
