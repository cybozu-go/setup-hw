package lib

import (
	"context"
	"os"
	"strings"

	"github.com/cybozu-go/well"
)

func CountCPUs() (int, error) {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return -1, err
	}
	info := strings.Split(string(data), "\n")
	count := map[string]struct{}{}
	for _, s := range info {
		if strings.Contains(s, "physical id") {
			count[s] = struct{}{}
		}
	}
	return len(count), nil
}

func CountMemoryModules(ctx context.Context) (int, error) {
	out, err := well.CommandContext(ctx, "dmidecode", "-t", "memory").Output()
	if err != nil {
		return -1, err
	}
	info := strings.Split(string(out), "\n")
	count := 0
	for _, s := range info {
		if strings.Contains(s, "Size: No Module Installed") {
			continue
		}
		ss := strings.TrimSpace(s)
		if strings.HasPrefix(ss, "Size:") {
			count++
		}
	}
	return count, nil
}
