package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/lib"
	"github.com/cybozu-go/well"
)

func main() {
	well.LogConfig{}.Apply()
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	if len(os.Args) < 2 {
		log.ErrorExit(fmt.Errorf("specify iso image file"))
	}

	vendor, err := lib.DetectVendor()
	if err != nil {
		log.ErrorExit(err)
	}

	var setup func(context.Context, string) error
	switch vendor {
	case lib.QEMU:
		setup = setupQEMU
	case lib.Dell:
		setup = setupDell
	default:
		log.ErrorExit(errors.New("unsupported vendor hardware"))
	}

	url := os.Args[1]

	err = setup(ctx, url)
	if err != nil {
		log.ErrorExit(err)
	}
}
