package main

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/lib"
	"github.com/cybozu-go/well"
)

func main() {
	well.LogConfig{}.Apply()
	ctx := context.Background()

	vendor, err := lib.DetectVendor()
	if err != nil {
		log.ErrorExit(err)
	}

	var setup func(context.Context, []string) error
	switch vendor {
	case lib.QEMU:
		setup = setupQEMU
	case lib.Dell:
		setup = setupDell
	default:
		log.ErrorExit(errors.New("unsupported vendor hardware"))
	}

	tmpdir, err := os.MkdirTemp("/tmp", "setup-apply-firmware-")
	if err != nil {
		log.ErrorExit(err)
	}
	defer os.RemoveAll(tmpdir)

	urls := os.Args[1:]
	files := make([]string, len(urls))
	for i, u := range urls {
		url, err := url.Parse(u)
		if err != nil {
			log.ErrorExit(err)
		}
		f := path.Base(url.Path)
		files[i] = filepath.Join(tmpdir, f)
	}

	err = downloadUpdaters(ctx, urls, files)
	if err != nil {
		log.ErrorExit(err)
	}

	err = setup(ctx, files)
	if err != nil {
		log.ErrorExit(err)
	}
}
