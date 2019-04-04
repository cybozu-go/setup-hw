package cmd

import (
	"testing"
)

func TestLogger(t *testing.T) {
	t.Parallel()

	logger := logger{}
	logger.Println("this should not fail", 42)
}
