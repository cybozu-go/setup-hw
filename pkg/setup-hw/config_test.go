package main

import (
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
}
