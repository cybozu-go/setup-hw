package main

import "context"

func monitorQEMU(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
