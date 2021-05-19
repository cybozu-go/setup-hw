package cmd

import (
	"context"
	"os"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
)

func monitorDell(ctx context.Context) error {
	if err := initDell(ctx); err != nil {
		return err
	}
	if err := resetDell(ctx); err != nil {
		return err
	}

	env := well.NewEnvironment(ctx)
	env.Go(func(ctx context.Context) error {
		for {
			select {
			case <-time.After(time.Duration(opts.resetInterval) * time.Hour):
			case <-ctx.Done():
				return nil
			}

			if _, err := os.Stat(opts.noResetFile); err == nil {
				// if no-reset file exists, skip reset.
				continue
			}

			if err := resetDell(ctx); err != nil {
				log.Error("failed to reset iDRAC", map[string]interface{}{
					log.FnError: err,
				})
				// continue working
			}
		}
	})

	env.Stop()
	return env.Wait()
}

func initDell(ctx context.Context) error {
	if err := well.CommandContext(ctx, "/usr/libexec/instsvcdrv-helper", "start").Run(); err != nil {
		return err
	}
	if err := well.CommandContext(ctx, "/opt/dell/srvadmin/bin/idracadm7", "remoteimage", "-d").Run(); err != nil {
		return err
	}
	return nil
}

func resetDell(ctx context.Context) error {
	return well.CommandContext(ctx, "/opt/dell/srvadmin/bin/idracadm7", "racreset", "soft").Run()
}
