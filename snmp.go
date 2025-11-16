package tplinkddm

import (
	"context"
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
func (c *SNMPClient) GetDDMMetrics(ctx context.Context) ([]DDMMetrics, error) {
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

	defer func() {
		if closeErr := client.Conn.Close(); closeErr != nil {
			slog.Debug("failed to close SNMP connection", "error", closeErr)
		}
	}()

	// Check context cancellation
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	// Walk the DDM status table
	ddmData, err := c.walkAllOIDs(client)
	if err != nil {
		return nil, err
	}

	// Parse and combine results
	return c.parseDDMMetrics(ddmData), nil
}

// walkOID performs SNMP walk and returns string values
func (c *SNMPClient) walkOID(client *gosnmp.GoSNMP, oid string) ([]string, error) {
	var results []string

	err := client.Walk(oid, func(pdu gosnmp.SnmpPDU) error {
		//nolint:exhaustive // Only OctetString is expected for TP-Link DDM DisplayString values
		switch pdu.Type {
		case gosnmp.OctetString:
			results = append(results, string(pdu.Value.([]byte)))
		default:
			// Other types are ignored
			slog.Debug("unexpected SNMP type", "oid", pdu.Name, "type", pdu.Type)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("SNMP walk failed: %w", err)
	}

	return results, nil
}

type ddmWalkData struct {
	ports        []string
	temps        []string
	voltages     []string
	biasCurrents []string
	txPowers     []string
	rxPowers     []string
}

func (c *SNMPClient) walkAllOIDs(client *gosnmp.GoSNMP) (*ddmWalkData, error) {
	data := &ddmWalkData{}

	var err error

	data.ports, err = c.walkOID(client, oidDDMStatusPort)
	if err != nil {
		return nil, fmt.Errorf("failed to get ports: %w", err)
	}

	data.temps, err = c.walkOID(client, oidDDMStatusTemperature)
	if err != nil {
		return nil, fmt.Errorf("failed to get temperatures: %w", err)
	}

	data.voltages, err = c.walkOID(client, oidDDMStatusVoltage)
	if err != nil {
		return nil, fmt.Errorf("failed to get voltages: %w", err)
	}

	data.biasCurrents, err = c.walkOID(client, oidDDMStatusBiasCurrent)
	if err != nil {
		return nil, fmt.Errorf("failed to get bias currents: %w", err)
	}

	data.txPowers, err = c.walkOID(client, oidDDMStatusTxPower)
	if err != nil {
		return nil, fmt.Errorf("failed to get TX powers: %w", err)
	}

	data.rxPowers, err = c.walkOID(client, oidDDMStatusRxPower)
	if err != nil {
		return nil, fmt.Errorf("failed to get RX powers: %w", err)
	}

	return data, nil
}

func (c *SNMPClient) parseDDMMetrics(data *ddmWalkData) []DDMMetrics {
	metrics := make([]DDMMetrics, 0, len(data.ports))

	for idx, portStr := range data.ports {
		port, err := parsePort(portStr)
		if err != nil {
			slog.Warn("skipping invalid port", "port", portStr, "error", err)

			continue
		}

		m := DDMMetrics{Port: port}

		if idx < len(data.temps) {
			m.Temperature, _ = parseFloat(data.temps[idx])
		}

		if idx < len(data.voltages) {
			m.Voltage, _ = parseFloat(data.voltages[idx])
		}

		if idx < len(data.biasCurrents) {
			m.BiasCurrent, _ = parseFloat(data.biasCurrents[idx])
		}

		if idx < len(data.txPowers) {
			m.TxPower, _ = parseFloat(data.txPowers[idx])
		}

		if idx < len(data.rxPowers) {
			m.RxPower, _ = parseFloat(data.rxPowers[idx])
		}

		metrics = append(metrics, m)
	}

	return metrics
}
