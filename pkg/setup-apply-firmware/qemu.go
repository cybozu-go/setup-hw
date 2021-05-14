package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
)

func setupQEMU(ctx context.Context, files []string) error {
	// nothing to do... just check file existence
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			return err
		}
		if info.Mode()&fs.ModeType != 0 {
			return fmt.Errorf("file %s is not a regular file", f)
		}
	}

	return nil
}
