package cmd

import (
	"context"

	"github.com/cybozu-go/well"
)

func initDell(ctx context.Context) error {
	if err := setupDell(ctx); err != nil {
		return err
	}
	return resetDell(ctx)
}

func setupDell(ctx context.Context) error {
	if err := well.CommandContext(ctx, "/usr/libexec/instsvcdrv-helper", "start").Run(); err != nil {
		return err
	}
	if err := well.CommandContext(ctx, "/usr/bin/racadm", "remoteimage", "-d").Run(); err != nil {
		return err
	}
	return nil
}

func resetDell(ctx context.Context) error {
	/*
		Depending on the F/W version, the option “soft” is required or an error occurs if given.
		Therefore, change the option and execute twice.
		This behavior has been verified at the link
		https://github.com/cybozu-private/neco-ops/issues/4113
	*/
	well.CommandContext(ctx, "/usr/bin/racadm", "racreset", "soft").Run()
	well.CommandContext(ctx, "/usr/bin/racadm", "racreset").Run()
	return nil
}
