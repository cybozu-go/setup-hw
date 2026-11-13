package main

import (
	"slices"
	"testing"

	"github.com/cybozu-go/setup-hw/lib"
)

func TestParseRacadmGetOutput(t *testing.T) {
	t.Parallel()

	val, err := parseRacadmGetOutput(`[Key=iDRAC.Embedded.1#SNMP.1]
	AgentEnable=Enabled
	
	`, "iDRAC.SNMP.AgentEnable")
	if err != nil {
		t.Fatal(err)
	}
	if val != "Enabled" {
		t.Error("unexpected value:", val)
	}

	val, err = parseRacadmGetOutput(`Enabled
	
	`, "System.ServerPwr.PSRapidOn")
	if err != nil {
		t.Fatal(err)
	}
	if val != "Enabled" {
		t.Error("unexpected value:", val)
	}
}

func TestParseRacadmGetOutputReadOnly(t *testing.T) {
	t.Parallel()

	// iDRAC10 keeps the '#' of a read-only attribute even when getting it alone.
	val, err := parseRacadmGetOutput(`[Key=BIOS.Setup.1-1#SysSecurity]
#TpmInfo=Type: 2.0  NTC

`, "BIOS.SysSecurity.TpmInfo")
	if err != nil {
		t.Fatal(err)
	}
	if val != "Type: 2.0  NTC" {
		t.Error("unexpected value:", val)
	}

	// iDRAC9 removes the '#'.
	val, err = parseRacadmGetOutput(`[Key=BIOS.Setup.1-1#SysSecurity]
TpmInfo=Type: 2.0  NTC

`, "BIOS.SysSecurity.TpmInfo")
	if err != nil {
		t.Fatal(err)
	}
	if val != "Type: 2.0  NTC" {
		t.Error("unexpected value:", val)
	}

	// Only the requested attribute is uncommented.
	out := uncommentReadOnly(`[Key=iDRAC.Embedded.1#NIC.1]
#ActiveNIC=Dedicated
SelectionByFQDD=Dedicated
`, "iDRAC.NIC.SelectionByFQDD")
	vals, err := parseRacadmGetOutputs([]byte(out), "iDRAC.NIC.SelectionByFQDD")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := vals["ActiveNIC"]; ok {
		t.Error("ActiveNIC should stay a comment:", vals)
	}
	if vals["SelectionByFQDD"] != "Dedicated" {
		t.Error("unexpected values:", vals)
	}
}

func TestWaitKeys(t *testing.T) {
	t.Parallel()

	gen16Only := []string{"System.ServerPwr.PSRapidOn", "iDRAC.NIC.Selection", "iDRAC.NIC.DNSRacName", "iDRAC.VirtualConsole.PluginType"}
	gen17Only := []string{"iDRAC.NIC.SelectionByFQDD", "iDRAC.Network.DNSRacName"}

	for _, p := range []lib.Product{lib.R6615, lib.R7615, lib.R6625} {
		keys := waitKeys(p)
		for _, k := range gen16Only {
			if !slices.Contains(keys, k) {
				t.Errorf("wait keys for %v should contain %s", p, k)
			}
		}
		for _, k := range gen17Only {
			if slices.Contains(keys, k) {
				t.Errorf("wait keys for %v should not contain %s", p, k)
			}
		}
	}

	for _, p := range []lib.Product{lib.R6715, lib.R5715} {
		keys := waitKeys(p)
		for _, k := range gen16Only {
			if slices.Contains(keys, k) {
				t.Errorf("wait keys for %v should not contain %s", p, k)
			}
		}
		for _, k := range gen17Only {
			if !slices.Contains(keys, k) {
				t.Errorf("wait keys for %v should contain %s", p, k)
			}
		}
	}
}
