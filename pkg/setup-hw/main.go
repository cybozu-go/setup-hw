package main

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/lib"
	"github.com/cybozu-go/well"
)

const (
	// AddressFile is the filename of the BMC address configuration file.
	AddressFile = "/etc/neco/bmc-address.json"

	// UserFile is the filename of the BMC user credentials.
	UserFile = "/etc/neco/bmc-user.json"

	// ExitReboot is the status code to tell the caller to reboot.
	ExitReboot = 10
)

func loadConfig() (*AddressConfig, *UserConfig, error) {
	f, err := os.Open(AddressFile)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	bmcAddress := new(AddressConfig)
	err = json.NewDecoder(f).Decode(bmcAddress)
	if err != nil {
		return nil, nil, err
	}

	if err := bmcAddress.Validate(); err != nil {
		return nil, nil, err
	}

	g, err := os.Open(UserFile)
	if err != nil {
		return nil, nil, err
	}
	defer g.Close()

	bmcUsers := new(UserConfig)
	err = json.NewDecoder(g).Decode(bmcUsers)
	if err != nil {
		return nil, nil, err
	}

	if err := bmcUsers.Validate(); err != nil {
		return nil, nil, err
	}

	return bmcAddress, bmcUsers, nil
}

func main() {
	well.LogConfig{}.Apply()

	ac, uc, err := loadConfig()
	if err != nil {
		log.ErrorExit(err)
	}

	vendor, err := lib.DetectVendor()
	if err != nil {
		log.ErrorExit(err)
	}

	var setup func(*AddressConfig, *UserConfig) (bool, error)
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
