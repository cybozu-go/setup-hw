package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
)

func setupDell(ctx context.Context, files []string) error {
	for _, f := range files {
		cmd := well.CommandContext(ctx, "/usr/bin/racadm", "update", "-f", f)
		buf := bytes.Buffer{}
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err := cmd.Run()
		// we cannot use exit status to detect errors because `racadm update` returns nonzero status even in case of successful update initiation.
		var exitError *exec.ExitError
		if err != nil && !errors.As(err, &exitError) {
			return fmt.Errorf("racadm update failed at file %s: %w", f, err)
		}
		msg := buf.String()
		if err = checkRacadmOutput(msg, f); err != nil {
			return err
		}
		log.Info("racadm update succeeded", map[string]interface{}{
			"file": f,
		})

		// if the next `racadm update` is executed immediately after the previous one, it will fail.
		time.Sleep(time.Second * 10)
	}

	return nil
}

func checkRacadmOutput(msg, f string) error {
	if !strings.Contains(msg, "\nRAC987: ") {
		return fmt.Errorf("racadm update failed at file %s: msg: %s", f, msg)
	}
	return nil
}
