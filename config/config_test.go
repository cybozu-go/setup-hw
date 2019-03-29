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

	bad := "8a8a8339"
	good := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC+CayRlLvi+VPdtAqVdq8PH/y/56XivHJ20QVggbxsxHt2CfGDc/gH9Nv9b/3lGuOQDh30TDX51Mj0JQEHRc3V7CGicOR2f+5fn2+8tvU2UrjBGab/SD2IyuSrsWr2qY4pIngGm8uE8x0lNd++1KZB+JxsJFUEG15ezLQV3WZWBIjjZtmvrtahvRApLlj9v55uvMzf3E8KTE/c1NDz3UxKq88e23ebBceXEnB7E1I0crDWN++L0ZivOIZ973b42s4hv1a4f6bV4xBCgMk4jqWCU8RJhFyLpdxQXgje1eQKegjVlDSe0xfeIzwJ1zXqEfOSh74lk7vzvQCk0AcdypiT dummy@dummy"

	goodCred := Credentials{
		AuthorizedKeys: []string{good},
	}
	badCred := Credentials{
		AuthorizedKeys: []string{good, bad},
	}

	uc := UserConfig{}
	if err := uc.Validate(); err != nil {
		t.Error(err)
	}
	uc.Root = goodCred
	if err := uc.Validate(); err != nil {
		t.Error(err)
	}
	uc.Root = badCred
	if err := uc.Validate(); err == nil {
		t.Error("credential for root must be invalid")
	}
	uc.Root = goodCred
	uc.Power = badCred
	if err := uc.Validate(); err == nil {
		t.Error("credential for power must be invalid")
	}
	uc.Power = goodCred
	uc.Support = badCred
	if err := uc.Validate(); err == nil {
		t.Error("credential for support must be invalid")
	}

	f, err := os.Open("../testdata/bmc-user.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	uc = UserConfig{}
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
	if uc.Power.Password.Raw != "ranranran" {
		t.Error("wrong power password")
	}
	if len(uc.Power.AuthorizedKeys) != 1 {
		t.Error("wrong power authorized keys")
	}
	if uc.Support.Password.Raw != "no support" {
		t.Error("wrong support password")
	}
}
