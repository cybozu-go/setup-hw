package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/cybozu-go/well"
)

const idracadm7Path = "/usr/bin/racadm"

func setupDell(ctx context.Context, url string) error {
	err := well.CommandContext(ctx, idracadm7Path, "vmdisconnect").Run()
	// we cannot use exit status to detect errors because `idracadm7 vmdisconnect` returns nonzero status if it is not connected.
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		return fmt.Errorf("racadm vmdisconnect failed: %w", err)
	}

	// `idracadm7 remoteimage -d` returns zero if the remote image is not connected.
	err = well.CommandContext(ctx, idracadm7Path, "remoteimage", "-d").Run()
	if err != nil {
		return fmt.Errorf("racadm remoteimage -d failed: %w", err)
	}

	// if `idracadm7 remoteimage -c` is executed immediately after disconnecting, it will fail.
	time.Sleep(time.Second * 5)

	err = well.CommandContext(ctx, idracadm7Path, "remoteimage", "-c", "-l", url).Run()
	if err != nil {
		return fmt.Errorf("racadm remoteimage -c failed: %w", err)
	}

	err = well.CommandContext(ctx, idracadm7Path, "set", "iDRAC.VirtualMedia.BootOnce", "1").Run()
	if err != nil {
		return fmt.Errorf("racadm set BootOnce failed: %w", err)
	}

	err = well.CommandContext(ctx, idracadm7Path, "set", "iDRAC.ServerBoot.FirstBootDevice", "VCD-DVD").Run()
	if err != nil {
		return fmt.Errorf("racadm set FirstBootDevice failed: %w", err)
	}

	return nil
}
