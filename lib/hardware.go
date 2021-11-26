package lib

import (
	"context"
	"os"
	"strings"

	"github.com/cybozu-go/well"
)

func DetectCPUNumber() (int, error) {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return -1, err
	}
	info := strings.Split(string(data), "\n")
	count := map[string]struct{}{}
	for _, s := range info {
		if !strings.Contains(s, "physical id") {
			continue
		}
		if _, ok := count[s]; !ok {
			count[s] = struct{}{}
		}
	}
	return len(count), nil
}

func DetectMemoryNumber(ctx context.Context) (int, error) {
	checkCmd := well.CommandContext(ctx, "dmidecode", "-t", "memory")
	out, err := checkCmd.Output()
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
