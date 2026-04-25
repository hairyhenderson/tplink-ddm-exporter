package tplinkddm

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/gosnmp/gosnmp"
)

// SNMP OIDs
const (
	// Standard MIB-II OIDs
	oidSysName = "1.3.6.1.2.1.1.5.0"

	// DDM Config (table .1.1.1.1)
	// oidDDMConfigPort     = "1.3.6.1.4.1.11863.6.96.1.1.1.1.1"
	oidDDMConfigStatus   = "1.3.6.1.4.1.11863.6.96.1.1.1.1.2" // 0=disable, 1=enable
	oidDDMConfigShutdown = "1.3.6.1.4.1.11863.6.96.1.1.1.1.3" // 0=none, 1=warning, 2=alarm
	oidDDMConfigPortLAG  = "1.3.6.1.4.1.11863.6.96.1.1.1.1.4" // LAG membership

	// DDM Status - Current values (table .1.7.1.1)
	oidDDMStatusPort        = "1.3.6.1.4.1.11863.6.96.1.7.1.1.1"
	oidDDMStatusTemperature = "1.3.6.1.4.1.11863.6.96.1.7.1.1.2"
	oidDDMStatusVoltage     = "1.3.6.1.4.1.11863.6.96.1.7.1.1.3"
	oidDDMStatusBiasCurrent = "1.3.6.1.4.1.11863.6.96.1.7.1.1.4"
	oidDDMStatusTxPower     = "1.3.6.1.4.1.11863.6.96.1.7.1.1.5"
	oidDDMStatusRxPower     = "1.3.6.1.4.1.11863.6.96.1.7.1.1.6"
	oidDDMStatusSupported   = "1.3.6.1.4.1.11863.6.96.1.7.1.1.7" // DDM supported
	// oidDDMStatusDataReady   = "1.3.6.1.4.1.11863.6.96.1.7.1.1.7" // SFP operational status (same as above, TP-Link quirk)
	oidDDMStatusLossSignal = "1.3.6.1.4.1.11863.6.96.1.7.1.1.8" // Loss of Signal (LOS)
	oidDDMStatusTxFault    = "1.3.6.1.4.1.11863.6.96.1.7.1.1.9" // Transmitter fault

	// RX Power thresholds (table .1.2.1.1)
	oidDDMRxPowerHighAlarm   = "1.3.6.1.4.1.11863.6.96.1.2.1.1.2"
	oidDDMRxPowerLowAlarm    = "1.3.6.1.4.1.11863.6.96.1.2.1.1.3"
	oidDDMRxPowerHighWarning = "1.3.6.1.4.1.11863.6.96.1.2.1.1.4"
	oidDDMRxPowerLowWarning  = "1.3.6.1.4.1.11863.6.96.1.2.1.1.5"

	// Voltage thresholds (table .1.3.1.1)
	oidDDMVoltageHighAlarm   = "1.3.6.1.4.1.11863.6.96.1.3.1.1.2"
	oidDDMVoltageLowAlarm    = "1.3.6.1.4.1.11863.6.96.1.3.1.1.3"
	oidDDMVoltageHighWarning = "1.3.6.1.4.1.11863.6.96.1.3.1.1.4"
	oidDDMVoltageLowWarning  = "1.3.6.1.4.1.11863.6.96.1.3.1.1.5"

	// Bias Current thresholds (table .1.4.1.1)
	oidDDMBiasCurrentHighAlarm   = "1.3.6.1.4.1.11863.6.96.1.4.1.1.2"
	oidDDMBiasCurrentLowAlarm    = "1.3.6.1.4.1.11863.6.96.1.4.1.1.3"
	oidDDMBiasCurrentHighWarning = "1.3.6.1.4.1.11863.6.96.1.4.1.1.4"
	oidDDMBiasCurrentLowWarning  = "1.3.6.1.4.1.11863.6.96.1.4.1.1.5"

	// TX Power thresholds (table .1.5.1.1)
	oidDDMTxPowerHighAlarm   = "1.3.6.1.4.1.11863.6.96.1.5.1.1.2"
	oidDDMTxPowerLowAlarm    = "1.3.6.1.4.1.11863.6.96.1.5.1.1.3"
	oidDDMTxPowerHighWarning = "1.3.6.1.4.1.11863.6.96.1.5.1.1.4"
	oidDDMTxPowerLowWarning  = "1.3.6.1.4.1.11863.6.96.1.5.1.1.5"

	// Temperature thresholds (table .1.6.1.1)
	oidDDMTemperatureHighAlarm   = "1.3.6.1.4.1.11863.6.96.1.6.1.1.2"
	oidDDMTemperatureLowAlarm    = "1.3.6.1.4.1.11863.6.96.1.6.1.1.3"
	oidDDMTemperatureHighWarning = "1.3.6.1.4.1.11863.6.96.1.6.1.1.4"
	oidDDMTemperatureLowWarning  = "1.3.6.1.4.1.11863.6.96.1.6.1.1.5"
)

// SNMPClient wraps gosnmp for TP-Link DDM queries
type SNMPClient struct {
	target    string
	community string
}

// DDMMetrics holds parsed DDM values for a port
type DDMMetrics struct {
	// Strings (16 bytes each on 64-bit)
	Port          string
	LAGMembership string // LAG/trunk membership (empty if not in LAG)

	// Float64 values (8 bytes each)
	Temperature float64
	Voltage     float64
	BiasCurrent float64
	TxPower     float64
	RxPower     float64

	// Temperature thresholds (Celsius)
	TemperatureHighAlarm   float64
	TemperatureLowAlarm    float64
	TemperatureHighWarning float64
	TemperatureLowWarning  float64

	// Voltage thresholds (Volts)
	VoltageHighAlarm   float64
	VoltageLowAlarm    float64
	VoltageHighWarning float64
	VoltageLowWarning  float64

	// Bias Current thresholds (mA)
	BiasCurrentHighAlarm   float64
	BiasCurrentLowAlarm    float64
	BiasCurrentHighWarning float64
	BiasCurrentLowWarning  float64

	// TX Power thresholds (dBm)
	TxPowerHighAlarm   float64
	TxPowerLowAlarm    float64
	TxPowerHighWarning float64
	TxPowerLowWarning  float64

	// RX Power thresholds (dBm)
	RxPowerHighAlarm   float64
	RxPowerLowAlarm    float64
	RxPowerHighWarning float64
	RxPowerLowWarning  float64

	// Int (8 bytes on 64-bit)
	ShutdownPolicy int // Port shutdown policy: 0=none, 1=warning, 2=alarm

	// Bools (1 byte each, but padded)
	DDMEnabled   bool // DDM monitoring enabled on port
	DDMSupported bool // SFP supports DDM
	LossOfSignal bool // SFP reports signal loss (LOS)
	TxFault      bool // SFP reports transmitter fault
}

// DDMResult holds the complete result of a DDM scrape
type DDMResult struct {
	SysName string
	Metrics []DDMMetrics
}

// NewSNMPClient creates a new SNMP client
func NewSNMPClient(target, community string) *SNMPClient {
	return &SNMPClient{
		target:    target,
		community: community,
	}
}

// GetDDMMetrics queries all DDM metrics and sysName from the switch
func (c *SNMPClient) GetDDMMetrics(ctx context.Context) (*DDMResult, error) {
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

	// Get sysName and walk the DDM status table
	ddmData, err := c.walkAllOIDs(client)
	if err != nil {
		return nil, err
	}

	// Parse and combine results
	return &DDMResult{
		SysName: ddmData.sysName,
		Metrics: c.parseDDMMetrics(ddmData),
	}, nil
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
	sysName string
	// Current values
	ports        []string
	temps        []string
	voltages     []string
	biasCurrents []string
	txPowers     []string
	rxPowers     []string

	// Configuration
	ddmEnabled     []string
	shutdownPolicy []string
	lagMembership  []string

	// Status flags
	ddmSupported []string
	lossOfSignal []string
	txFault      []string

	// Temperature thresholds
	tempHighAlarm   []string
	tempLowAlarm    []string
	tempHighWarning []string
	tempLowWarning  []string

	// Voltage thresholds
	voltageHighAlarm   []string
	voltageLowAlarm    []string
	voltageHighWarning []string
	voltageLowWarning  []string

	// Bias Current thresholds
	biasCurrentHighAlarm   []string
	biasCurrentLowAlarm    []string
	biasCurrentHighWarning []string
	biasCurrentLowWarning  []string

	// TX Power thresholds
	txPowerHighAlarm   []string
	txPowerLowAlarm    []string
	txPowerHighWarning []string
	txPowerLowWarning  []string

	// RX Power thresholds
	rxPowerHighAlarm   []string
	rxPowerLowAlarm    []string
	rxPowerHighWarning []string
	rxPowerLowWarning  []string
}

//nolint:gocyclo,funlen,gocognit // Walking many OIDs for DDM thresholds
func (c *SNMPClient) walkAllOIDs(client *gosnmp.GoSNMP) (*ddmWalkData, error) {
	data := &ddmWalkData{}

	var err error

	// Get sysName (single value, not a walk)
	result, err := client.Get([]string{oidSysName})
	if err != nil {
		slog.Debug("failed to get sysName", "error", err)
	} else if len(result.Variables) > 0 {
		pdu := result.Variables[0]
		if pdu.Type == gosnmp.OctetString {
			data.sysName = string(pdu.Value.([]byte))
		}
	}

	// Current values
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

	// Configuration
	data.ddmEnabled, err = c.walkOID(client, oidDDMConfigStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to get DDM enabled status: %w", err)
	}

	data.shutdownPolicy, err = c.walkOID(client, oidDDMConfigShutdown)
	if err != nil {
		return nil, fmt.Errorf("failed to get shutdown policy: %w", err)
	}

	data.lagMembership, err = c.walkOID(client, oidDDMConfigPortLAG)
	if err != nil {
		return nil, fmt.Errorf("failed to get LAG membership: %w", err)
	}

	// Status flags
	data.ddmSupported, err = c.walkOID(client, oidDDMStatusSupported)
	if err != nil {
		return nil, fmt.Errorf("failed to get DDM supported: %w", err)
	}

	data.lossOfSignal, err = c.walkOID(client, oidDDMStatusLossSignal)
	if err != nil {
		return nil, fmt.Errorf("failed to get loss of signal: %w", err)
	}

	data.txFault, err = c.walkOID(client, oidDDMStatusTxFault)
	if err != nil {
		return nil, fmt.Errorf("failed to get TX fault: %w", err)
	}

	// Temperature thresholds
	data.tempHighAlarm, err = c.walkOID(client, oidDDMTemperatureHighAlarm)
	if err != nil {
		return nil, fmt.Errorf("failed to get temperature high alarm: %w", err)
	}

	data.tempLowAlarm, err = c.walkOID(client, oidDDMTemperatureLowAlarm)
	if err != nil {
		return nil, fmt.Errorf("failed to get temperature low alarm: %w", err)
	}

	data.tempHighWarning, err = c.walkOID(client, oidDDMTemperatureHighWarning)
	if err != nil {
		return nil, fmt.Errorf("failed to get temperature high warning: %w", err)
	}

	data.tempLowWarning, err = c.walkOID(client, oidDDMTemperatureLowWarning)
	if err != nil {
		return nil, fmt.Errorf("failed to get temperature low warning: %w", err)
	}

	// Voltage thresholds
	data.voltageHighAlarm, err = c.walkOID(client, oidDDMVoltageHighAlarm)
	if err != nil {
		return nil, fmt.Errorf("failed to get voltage high alarm: %w", err)
	}

	data.voltageLowAlarm, err = c.walkOID(client, oidDDMVoltageLowAlarm)
	if err != nil {
		return nil, fmt.Errorf("failed to get voltage low alarm: %w", err)
	}

	data.voltageHighWarning, err = c.walkOID(client, oidDDMVoltageHighWarning)
	if err != nil {
		return nil, fmt.Errorf("failed to get voltage high warning: %w", err)
	}

	data.voltageLowWarning, err = c.walkOID(client, oidDDMVoltageLowWarning)
	if err != nil {
		return nil, fmt.Errorf("failed to get voltage low warning: %w", err)
	}

	// Bias Current thresholds
	data.biasCurrentHighAlarm, err = c.walkOID(client, oidDDMBiasCurrentHighAlarm)
	if err != nil {
		return nil, fmt.Errorf("failed to get bias current high alarm: %w", err)
	}

	data.biasCurrentLowAlarm, err = c.walkOID(client, oidDDMBiasCurrentLowAlarm)
	if err != nil {
		return nil, fmt.Errorf("failed to get bias current low alarm: %w", err)
	}

	data.biasCurrentHighWarning, err = c.walkOID(client, oidDDMBiasCurrentHighWarning)
	if err != nil {
		return nil, fmt.Errorf("failed to get bias current high warning: %w", err)
	}

	data.biasCurrentLowWarning, err = c.walkOID(client, oidDDMBiasCurrentLowWarning)
	if err != nil {
		return nil, fmt.Errorf("failed to get bias current low warning: %w", err)
	}

	// TX Power thresholds
	data.txPowerHighAlarm, err = c.walkOID(client, oidDDMTxPowerHighAlarm)
	if err != nil {
		return nil, fmt.Errorf("failed to get TX power high alarm: %w", err)
	}

	data.txPowerLowAlarm, err = c.walkOID(client, oidDDMTxPowerLowAlarm)
	if err != nil {
		return nil, fmt.Errorf("failed to get TX power low alarm: %w", err)
	}

	data.txPowerHighWarning, err = c.walkOID(client, oidDDMTxPowerHighWarning)
	if err != nil {
		return nil, fmt.Errorf("failed to get TX power high warning: %w", err)
	}

	data.txPowerLowWarning, err = c.walkOID(client, oidDDMTxPowerLowWarning)
	if err != nil {
		return nil, fmt.Errorf("failed to get TX power low warning: %w", err)
	}

	// RX Power thresholds
	data.rxPowerHighAlarm, err = c.walkOID(client, oidDDMRxPowerHighAlarm)
	if err != nil {
		return nil, fmt.Errorf("failed to get RX power high alarm: %w", err)
	}

	data.rxPowerLowAlarm, err = c.walkOID(client, oidDDMRxPowerLowAlarm)
	if err != nil {
		return nil, fmt.Errorf("failed to get RX power low alarm: %w", err)
	}

	data.rxPowerHighWarning, err = c.walkOID(client, oidDDMRxPowerHighWarning)
	if err != nil {
		return nil, fmt.Errorf("failed to get RX power high warning: %w", err)
	}

	data.rxPowerLowWarning, err = c.walkOID(client, oidDDMRxPowerLowWarning)
	if err != nil {
		return nil, fmt.Errorf("failed to get RX power low warning: %w", err)
	}

	return data, nil
}

//nolint:gocognit,gocyclo,funlen // Parsing many DDM threshold fields
func (c *SNMPClient) parseDDMMetrics(data *ddmWalkData) []DDMMetrics {
	metrics := make([]DDMMetrics, 0, len(data.ports))

	for idx, portStr := range data.ports {
		port, err := parsePort(portStr)
		if err != nil {
			slog.Warn("skipping invalid port", "port", portStr, "error", err)

			continue
		}

		m := DDMMetrics{Port: port}

		// Current values
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

		// Configuration
		if idx < len(data.ddmEnabled) {
			// DDM config uses 0=disable, 1=enable (same as boolean)
			m.DDMEnabled, _ = strconv.ParseBool(data.ddmEnabled[idx])
		}

		if idx < len(data.shutdownPolicy) {
			m.ShutdownPolicy, _ = strconv.Atoi(data.shutdownPolicy[idx])
		}

		if idx < len(data.lagMembership) {
			m.LAGMembership = data.lagMembership[idx]
		}

		// Status flags
		if idx < len(data.ddmSupported) {
			m.DDMSupported, _ = strconv.ParseBool(data.ddmSupported[idx])
		}

		if idx < len(data.lossOfSignal) {
			m.LossOfSignal, _ = strconv.ParseBool(data.lossOfSignal[idx])
		}

		if idx < len(data.txFault) {
			m.TxFault, _ = strconv.ParseBool(data.txFault[idx])
		}

		// Temperature thresholds
		if idx < len(data.tempHighAlarm) {
			m.TemperatureHighAlarm, _ = parseFloat(data.tempHighAlarm[idx])
		}

		if idx < len(data.tempLowAlarm) {
			m.TemperatureLowAlarm, _ = parseFloat(data.tempLowAlarm[idx])
		}

		if idx < len(data.tempHighWarning) {
			m.TemperatureHighWarning, _ = parseFloat(data.tempHighWarning[idx])
		}

		if idx < len(data.tempLowWarning) {
			m.TemperatureLowWarning, _ = parseFloat(data.tempLowWarning[idx])
		}

		// Voltage thresholds
		if idx < len(data.voltageHighAlarm) {
			m.VoltageHighAlarm, _ = parseFloat(data.voltageHighAlarm[idx])
		}

		if idx < len(data.voltageLowAlarm) {
			m.VoltageLowAlarm, _ = parseFloat(data.voltageLowAlarm[idx])
		}

		if idx < len(data.voltageHighWarning) {
			m.VoltageHighWarning, _ = parseFloat(data.voltageHighWarning[idx])
		}

		if idx < len(data.voltageLowWarning) {
			m.VoltageLowWarning, _ = parseFloat(data.voltageLowWarning[idx])
		}

		// Bias Current thresholds
		if idx < len(data.biasCurrentHighAlarm) {
			m.BiasCurrentHighAlarm, _ = parseFloat(data.biasCurrentHighAlarm[idx])
		}

		if idx < len(data.biasCurrentLowAlarm) {
			m.BiasCurrentLowAlarm, _ = parseFloat(data.biasCurrentLowAlarm[idx])
		}

		if idx < len(data.biasCurrentHighWarning) {
			m.BiasCurrentHighWarning, _ = parseFloat(data.biasCurrentHighWarning[idx])
		}

		if idx < len(data.biasCurrentLowWarning) {
			m.BiasCurrentLowWarning, _ = parseFloat(data.biasCurrentLowWarning[idx])
		}

		// TX Power thresholds
		if idx < len(data.txPowerHighAlarm) {
			m.TxPowerHighAlarm, _ = parseFloat(data.txPowerHighAlarm[idx])
		}

		if idx < len(data.txPowerLowAlarm) {
			m.TxPowerLowAlarm, _ = parseFloat(data.txPowerLowAlarm[idx])
		}

		if idx < len(data.txPowerHighWarning) {
			m.TxPowerHighWarning, _ = parseFloat(data.txPowerHighWarning[idx])
		}

		if idx < len(data.txPowerLowWarning) {
			m.TxPowerLowWarning, _ = parseFloat(data.txPowerLowWarning[idx])
		}

		// RX Power thresholds
		if idx < len(data.rxPowerHighAlarm) {
			m.RxPowerHighAlarm, _ = parseFloat(data.rxPowerHighAlarm[idx])
		}

		if idx < len(data.rxPowerLowAlarm) {
			m.RxPowerLowAlarm, _ = parseFloat(data.rxPowerLowAlarm[idx])
		}

		if idx < len(data.rxPowerHighWarning) {
			m.RxPowerHighWarning, _ = parseFloat(data.rxPowerHighWarning[idx])
		}

		if idx < len(data.rxPowerLowWarning) {
			m.RxPowerLowWarning, _ = parseFloat(data.rxPowerLowWarning[idx])
		}

		metrics = append(metrics, m)
	}

	return metrics
}
