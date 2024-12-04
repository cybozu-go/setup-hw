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

// Product represents server hardware product name.
type Product int

// Product
const (
	Other Product = iota
	R6525
	R7525
	R6615
	R7615
)

// DetectProduct detects product name.
func DetectProduct() (Product, error) {
	data, err := os.ReadFile("/sys/devices/virtual/dmi/id/product_name")
	if err != nil {
		return Other, err
	}
	product := strings.TrimSpace(string(data))
	switch {
	case strings.Contains(product, "R6525"):
		return R6525, nil
	case strings.Contains(product, "R7525"):
		return R7525, nil
	case strings.Contains(product, "R6615"):
		return R6615, nil
	case strings.Contains(product, "R7615"):
		return R7615, nil
	default:
		return Other, nil
	}
}
