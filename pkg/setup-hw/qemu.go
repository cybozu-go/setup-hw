package main

import (
	"os"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/config"
)

const virtualBMCPort = "/dev/virtio-ports/placemat"

// setupQEMU configures virtual BMC provided by placemat.
// https://github.com/cybozu-go/placemat/blob/master/docs/virtual_bmc.md
func setupQEMU(ac *config.AddressConfig, uc *config.UserConfig) (bool, error) {
	f, err := os.OpenFile(virtualBMCPort, os.O_WRONLY, 0644)
	if err == nil {
		_, err = f.WriteString(ac.IPv4.Address + "\n")
		f.Close()
		return false, err
	}

	if os.IsNotExist(err) {
		log.Warn("virtual BMC is not found", nil)
		return false, nil
	}
	return false, err
}
