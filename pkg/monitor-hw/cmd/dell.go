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
	return well.CommandContext(ctx, "/usr/bin/racadm", "racreset", "soft").Run()
}
