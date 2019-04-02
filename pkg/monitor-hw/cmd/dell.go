package cmd

import (
	"context"
	"time"

	"github.com/cybozu-go/well"
)

func monitorDell(ctx context.Context) error {
	if err := initDell(ctx); err != nil {
		return err
	}

	env := well.NewEnvironment(ctx)
	env.Go(func(ctx context.Context) error {
		for {
			select {
			case <-time.After(time.Duration(opts.resetInterval) * time.Second):
			case <-ctx.Done():
				return nil
			}
			// TODO: reset iDRAC
		}
	})

	env.Stop()
	return env.Wait()
}

func initDell(ctx context.Context) error {
	// TODO: uncomment this
	//return well.CommandContext(ctx, "/usr/libexec/instsvcdrv-helper", "start").Run()
	return nil
}
