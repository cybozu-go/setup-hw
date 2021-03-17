package lib

import (
	"errors"
	"os"
	"strings"
)

// Vendor represents server hardware vendor.
type Vendor int

// Vendors
const (
	Unknown Vendor = iota
	QEMU
	Dell
)

// DetectVendor detects hardware vendor.
func DetectVendor() (Vendor, error) {
	data, err := os.ReadFile("/sys/devices/virtual/dmi/id/sys_vendor")
	if err != nil {
		return Unknown, err
	}

	vendor := strings.TrimSpace(string(data))
	if vendor == "QEMU" {
		return QEMU, nil
	}

	if strings.HasPrefix(vendor, "Dell") {
		return Dell, nil
	}

	return Unknown, errors.New("unknown vendor: " + vendor)
}
