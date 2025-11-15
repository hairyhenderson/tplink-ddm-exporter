package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gosnmp/gosnmp"
)

// TP-Link DDM Status MIB OIDs
const (
	oidDDMStatusPort        = "1.3.6.1.4.1.11863.6.96.1.7.1.1.1"
	oidDDMStatusTemperature = "1.3.6.1.4.1.11863.6.96.1.7.1.1.2"
	oidDDMStatusVoltage     = "1.3.6.1.4.1.11863.6.96.1.7.1.1.3"
	oidDDMStatusBiasCurrent = "1.3.6.1.4.1.11863.6.96.1.7.1.1.4"
	oidDDMStatusTxPower     = "1.3.6.1.4.1.11863.6.96.1.7.1.1.5"
	oidDDMStatusRxPower     = "1.3.6.1.4.1.11863.6.96.1.7.1.1.6"
)

// SNMPClient wraps gosnmp for TP-Link DDM queries
type SNMPClient struct {
	target    string
	community string
}

// DDMMetrics holds parsed DDM values for a port
type DDMMetrics struct {
	Port        string
	Temperature float64
	Voltage     float64
	BiasCurrent float64
	TxPower     float64
	RxPower     float64
}

// NewSNMPClient creates a new SNMP client
func NewSNMPClient(target, community string) *SNMPClient {
	return &SNMPClient{
		target:    target,
		community: community,
	}
}

// GetDDMMetrics queries all DDM metrics from the switch
func (c *SNMPClient) GetDDMMetrics() ([]DDMMetrics, error) {
	// Create gosnmp client
	client := &gosnmp.GoSNMP{
		Target:    c.target,
		Port:      161,
		Community: c.community,
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(10) * time.Second,
		Retries:   3,
	}

	err := client.Connect()
	if err != nil {
		return nil, fmt.Errorf("SNMP connect failed: %w", err)
	}
	defer client.Conn.Close()

	// Walk the DDM status table
	ports, err := c.walkOID(client, oidDDMStatusPort)
	if err != nil {
		return nil, fmt.Errorf("failed to get ports: %w", err)
	}

	temps, err := c.walkOID(client, oidDDMStatusTemperature)
	if err != nil {
		return nil, fmt.Errorf("failed to get temperatures: %w", err)
	}

	voltages, err := c.walkOID(client, oidDDMStatusVoltage)
	if err != nil {
		return nil, fmt.Errorf("failed to get voltages: %w", err)
	}

	biasCurrents, err := c.walkOID(client, oidDDMStatusBiasCurrent)
	if err != nil {
		return nil, fmt.Errorf("failed to get bias currents: %w", err)
	}

	txPowers, err := c.walkOID(client, oidDDMStatusTxPower)
	if err != nil {
		return nil, fmt.Errorf("failed to get TX powers: %w", err)
	}

	rxPowers, err := c.walkOID(client, oidDDMStatusRxPower)
	if err != nil {
		return nil, fmt.Errorf("failed to get RX powers: %w", err)
	}

	// Parse and combine results
	var metrics []DDMMetrics
	for idx, portStr := range ports {
		port, err := parsePort(portStr)
		if err != nil {
			slog.Warn("skipping invalid port", "port", portStr, "error", err)
			continue
		}

		m := DDMMetrics{Port: port}

		if idx < len(temps) {
			m.Temperature, _ = parseFloat(temps[idx])
		}
		if idx < len(voltages) {
			m.Voltage, _ = parseFloat(voltages[idx])
		}
		if idx < len(biasCurrents) {
			m.BiasCurrent, _ = parseFloat(biasCurrents[idx])
		}
		if idx < len(txPowers) {
			m.TxPower, _ = parseFloat(txPowers[idx])
		}
		if idx < len(rxPowers) {
			m.RxPower, _ = parseFloat(rxPowers[idx])
		}

		metrics = append(metrics, m)
	}

	return metrics, nil
}

// walkOID performs SNMP walk and returns string values
func (c *SNMPClient) walkOID(client *gosnmp.GoSNMP, oid string) ([]string, error) {
	var results []string

	err := client.Walk(oid, func(pdu gosnmp.SnmpPDU) error {
		switch pdu.Type {
		case gosnmp.OctetString:
			results = append(results, string(pdu.Value.([]byte)))
		default:
			slog.Debug("unexpected SNMP type", "oid", pdu.Name, "type", pdu.Type)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

