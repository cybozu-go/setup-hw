package main

import (
	"errors"
	"net"

	"golang.org/x/crypto/ssh"
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
	Password       BMCPassword `json:"password"`
	AuthorizedKeys []string    `json:"authorized_keys"`
}

// Validate validates Credentials.
func (c Credentials) Validate() error {
	for _, k := range c.AuthorizedKeys {
		_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(k))
		if err != nil {
			return errors.New("invalid authorized key: " + k)
		}
	}
	return nil
}

// UserConfig represents a set of BMC user credentials in JSON format.
type UserConfig struct {
	Root    Credentials `json:"root"`
	Power   Credentials `json:"power"`
	Support Credentials `json:"support"`
}

// Validate validates UserConfig.
func (c UserConfig) Validate() error {
	if err := c.Root.Validate(); err != nil {
		return err
	}
	if err := c.Power.Validate(); err != nil {
		return err
	}
	if err := c.Support.Validate(); err != nil {
		return err
	}
	return nil
}
