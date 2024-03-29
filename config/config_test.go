package config

import (
	"encoding/json"
	"os"
	"testing"
)

func TestIPv4Config(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Address string
		Netmask string
		Gateway string
		valid   bool
	}{
		{"10.1.2.3", "255.255.255.0", "10.1.2.1", true},
		{"10.1.2.3", "255.255.255.0", "10.1.3.1", false},
		{"10.1.2.3", "255.255.0.240", "10.1.0.1", false},
		{"abcd::1", "255.255.255.0", "10.1.2.1", false},
		{"10.1.2.3", "ffff::0", "10.1.2.1", false},
		{"10.1.2.3", "255.255.255.0", "abcd::1", false},
	}

	for _, tc := range tests {
		c := IPv4Config{tc.Address, tc.Netmask, tc.Gateway}
		err := c.Validate()
		if err != nil {
			if tc.valid {
				t.Error(err)
			}
			continue
		}

		if !tc.valid {
			t.Errorf("validation failed for %+v", c)
		}
	}
}

func TestAddressConfig(t *testing.T) {
	t.Parallel()

	f, err := os.Open("../testdata/bmc-address.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var c AddressConfig
	err = json.NewDecoder(f).Decode(&c)
	if err != nil {
		t.Fatal(err)
	}

	if c.IPv4.Address != "1.2.3.4" {
		t.Error(`c.IPv4.Address != "1.2.3.4"`)
	}
	if c.IPv4.Netmask != "255.255.255.0" {
		t.Error(`c.IPv4.Netmask != "255.255.255.0"`)
	}
	if c.IPv4.Gateway != "1.2.3.1" {
		t.Error(`c.IPv4.Gateway != "1.2.3.1"`)
	}
}

func TestUserConfig(t *testing.T) {
	t.Parallel()

	f, err := os.Open("../testdata/bmc-user.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	uc := UserConfig{}
	err = json.NewDecoder(f).Decode(&uc)
	if err != nil {
		t.Fatal(err)
	}

	if uc.Root.Password.Hash != "7DC9C71A51E0B494700E52CC11F80E2837C025D60120EE8C6F03D5A3C91B504A" {
		t.Error("wrong root password hash")
	}
	if uc.Root.Password.Salt != "593C31FF6D409480F032AA2FF6EC781E" {
		t.Error("wrong root password salt")
	}
	if uc.Repair.Password.Hash != "78B13CF445B376D74BD1BEBA0B8802AD691D69483E752191D06D3C0AF362DAD8" {
		t.Error("wrong repair password hash")
	}
	if uc.Repair.Password.Salt != "8E4934DDBEEE7C9AE9427A8283D7FA10" {
		t.Error("wrong repair password salt")
	}
	if uc.Power.Password.Raw != "ranranran" {
		t.Error("wrong power password")
	}
	if uc.Support.Password.Raw != "no support" {
		t.Error("wrong support password")
	}
}
