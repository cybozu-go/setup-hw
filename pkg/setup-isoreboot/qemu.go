package main

import (
	"context"

	"github.com/cybozu-go/log"
)

func setupQEMU(ctx context.Context, url string) error {
	// nothing to do

	log.Info("setupQEMU called", map[string]interface{}{
		"url": url,
	})

	return nil
}
