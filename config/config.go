package config

import (
	"encoding/json"
	"errors"
	"net"
	"os"
)

const (
	// AddressFile is the filename of the BMC address configuration file.
	AddressFile = "/etc/neco/bmc-address.json"

	// UserFile is the filename of the BMC user credentials.
	UserFile = "/etc/neco/bmc-user.json"
)

// IPv4Config represents NIC configuration parameters for IPv4 network.
type IPv4Config struct {
	Address string `json:"address"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
}

// Validate validates IPv4 configuration.
func (c IPv4Config) Validate() error {
	ip := net.ParseIP(c.Address).To4()
	if ip == nil {
		return errors.New("invalid address: " + c.Address)
	}

	mask := net.IPMask(net.ParseIP(c.Netmask).To4())
	if mask == nil {
		return errors.New("invalid netmask: " + c.Netmask)
	}
	ones, bits := mask.Size()
	if ones == 0 && bits == 0 {
		return errors.New("invalid netmask: " + c.Netmask)
	}

	n := &net.IPNet{
		IP:   ip.Mask(mask),
		Mask: mask,
	}

	gw := net.ParseIP(c.Gateway).To4()
	if gw == nil {
		return errors.New("invalid gateway: " + c.Gateway)
	}

	if !n.Contains(gw) {
		return errors.New("gateway address is out of the network: " + c.Gateway)
	}

	return nil
}

// AddressConfig represents BMC NIC configuration in JSON format.
type AddressConfig struct {
	IPv4 IPv4Config `json:"ipv4"`
}

// Validate validates address configuration.
func (c AddressConfig) Validate() error {
	return c.IPv4.Validate()
}

// BMCPassword represents password for a BMC user.
type BMCPassword struct {
	Raw  string `json:"raw"`
	Hash string `json:"hash"`
	Salt string `json:"salt"`
}

// Credentials represents credentials of a BMC user.
type Credentials struct {
	Password BMCPassword `json:"password"`
}

// UserConfig represents a set of BMC user credentials in JSON format.
type UserConfig struct {
	Root    Credentials `json:"root"`
	Power   Credentials `json:"power"`
	Support Credentials `json:"support"`
}

// LoadConfig loads AddressConfig and UserConfig.
func LoadConfig() (*AddressConfig, *UserConfig, error) {
	f, err := os.Open(AddressFile)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	bmcAddress := new(AddressConfig)
	err = json.NewDecoder(f).Decode(bmcAddress)
	if err != nil {
		return nil, nil, err
	}

	if err := bmcAddress.Validate(); err != nil {
		return nil, nil, err
	}

	g, err := os.Open(UserFile)
	if err != nil {
		return nil, nil, err
	}
	defer g.Close()

	bmcUsers := new(UserConfig)
	err = json.NewDecoder(g).Decode(bmcUsers)
	if err != nil {
		return nil, nil, err
	}

	return bmcAddress, bmcUsers, nil
}
