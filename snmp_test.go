package main

import (
	"testing"
)

func TestNewSNMPClient(t *testing.T) {
	client := NewSNMPClient("192.168.2.96", "public")
	if client == nil {
		t.Fatal("NewSNMPClient returned nil")
	}
	if client.target != "192.168.2.96" {
		t.Errorf("target = %v, want 192.168.2.96", client.target)
	}
	if client.community != "public" {
		t.Errorf("community = %v, want public", client.community)
	}
}

func TestOIDs(t *testing.T) {
	// Validate OID constants are correct
	expectedOIDs := map[string]string{
		"port":        "1.3.6.1.4.1.11863.6.96.1.7.1.1.1",
		"temperature": "1.3.6.1.4.1.11863.6.96.1.7.1.1.2",
		"voltage":     "1.3.6.1.4.1.11863.6.96.1.7.1.1.3",
		"biasCurrent": "1.3.6.1.4.1.11863.6.96.1.7.1.1.4",
		"txPower":     "1.3.6.1.4.1.11863.6.96.1.7.1.1.5",
		"rxPower":     "1.3.6.1.4.1.11863.6.96.1.7.1.1.6",
	}

	if oidDDMStatusPort != expectedOIDs["port"] {
		t.Errorf("oidDDMStatusPort = %v, want %v", oidDDMStatusPort, expectedOIDs["port"])
	}
	if oidDDMStatusTemperature != expectedOIDs["temperature"] {
		t.Errorf("oidDDMStatusTemperature = %v, want %v", oidDDMStatusTemperature, expectedOIDs["temperature"])
	}
	if oidDDMStatusVoltage != expectedOIDs["voltage"] {
		t.Errorf("oidDDMStatusVoltage = %v, want %v", oidDDMStatusVoltage, expectedOIDs["voltage"])
	}
	if oidDDMStatusBiasCurrent != expectedOIDs["biasCurrent"] {
		t.Errorf("oidDDMStatusBiasCurrent = %v, want %v", oidDDMStatusBiasCurrent, expectedOIDs["biasCurrent"])
	}
	if oidDDMStatusTxPower != expectedOIDs["txPower"] {
		t.Errorf("oidDDMStatusTxPower = %v, want %v", oidDDMStatusTxPower, expectedOIDs["txPower"])
	}
	if oidDDMStatusRxPower != expectedOIDs["rxPower"] {
		t.Errorf("oidDDMStatusRxPower = %v, want %v", oidDDMStatusRxPower, expectedOIDs["rxPower"])
	}
}
