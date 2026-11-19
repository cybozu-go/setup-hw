package cmd

import (
	"encoding/base64"
	"testing"
)

func TestHashPasswordIDRAC10(t *testing.T) {
	t.Parallel()

	// The expected value was read from iDRAC10 after setting the password
	// with "racadm set iDRAC.Users.N.Password".
	salt, err := base64.StdEncoding.DecodeString("V0GbeeYDOfcFDQFgGhEefFa++gs1snm6r9bHFFLiOGM=")
	if err != nil {
		t.Fatal(err)
	}
	got, err := hashPasswordIDRAC10([]byte("Test-Passw0rd-123"), salt)
	if err != nil {
		t.Fatal(err)
	}
	want := "LkbMFJ0RsHVNUVly44/bq6VPVzcfSB+T/BzL2YflGgM=|V0GbeeYDOfcFDQFgGhEefFa++gs1snm6r9bHFFLiOGM="
	if got != want {
		t.Errorf("unexpected hash: got %s, want %s", got, want)
	}
}
