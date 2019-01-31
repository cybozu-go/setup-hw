package main

import (
	"testing"
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
