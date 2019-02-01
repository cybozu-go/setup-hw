package main

import (
	"context"

	"github.com/cybozu-go/well"
)

func monitorDell(ctx context.Context) error {
	if err := initDell(ctx); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

func initDell(ctx context.Context) error {
	return well.CommandContext(ctx, "/usr/libexec/instsvcdrv-helper", "start").Run()
}
