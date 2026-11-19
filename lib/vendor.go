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
	R6625
	R6715
	R5715
)

// Is17G returns true if the product is a Dell 17th generation server.
// 17G servers are managed by iDRAC10, whose racadm attributes differ from iDRAC9.
// Support for 17G servers is preliminary and has not been fully verified yet.
func (p Product) Is17G() bool {
	switch p {
	case R6715, R5715:
		return true
	default:
		return false
	}
}

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
	case strings.Contains(product, "R6625"):
		return R6625, nil
	case strings.Contains(product, "R6715"):
		return R6715, nil
	case strings.Contains(product, "R5715"):
		return R5715, nil
	default:
		return Other, nil
	}
}
