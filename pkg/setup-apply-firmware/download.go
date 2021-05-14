package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cybozu-go/well"
)

func downloadUpdaters(ctx context.Context, urls, files []string) error {
	if len(urls) != len(files) {
		return fmt.Errorf("length of urls and files are defferent")
	}

	for i, u := range urls {
		f := files[i]
		cmd := well.CommandContext(ctx, "/usr/bin/curl", "-sSfo", f, u)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
