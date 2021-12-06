package lib

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/cybozu-go/well"
)

func CountCPUs() (int, error) {
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return -1, err
	}
	s := bufio.NewScanner(f)
	count := map[string]struct{}{}
	for s.Scan() {
		l := s.Text()
		if strings.Contains(l, "physical id") {
			count[l] = struct{}{}
		}
	}
	if err := s.Err(); err != nil {
		return -1, err
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
