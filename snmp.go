package tplinkddm

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// SNMP OIDs
const (
	// Standard MIB-II OIDs
	oidSysName = "1.3.6.1.2.1.1.5.0"

	// DDM root OID - walk this to get all DDM data in one BulkWalk
	oidDDMRoot = "1.3.6.1.4.1.11863.6.96.1"

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

//nolint:gochecknoglobals // package-level tracer is the OTel convention
var tracer = otel.Tracer("github.com/hairyhenderson/tplink-ddm-exporter")

// GetDDMMetrics queries all DDM metrics and sysName from the switch
func (c *SNMPClient) GetDDMMetrics(ctx context.Context) (*DDMResult, error) {
	ctx, span := tracer.Start(ctx, "SNMPClient.GetDDMMetrics",
		trace.WithAttributes(
			attribute.String("snmp.target", c.target),
		),
	)
	defer span.End()

	client := &gosnmp.GoSNMP{
		Target:    c.target,
		Port:      161,
		Community: c.community,
		Version:   gosnmp.Version2c,
		Timeout:   2 * time.Second,
		Retries:   1,
	}

	err := client.Connect()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "SNMP connect failed")

		return nil, fmt.Errorf("SNMP connect failed: %w", err)
	}

	defer func() {
		if closeErr := client.Conn.Close(); closeErr != nil {
			slog.Debug("failed to close SNMP connection", "error", closeErr)
		}
	}()

	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	ddmData, err := c.walkAllOIDs(ctx, client)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "walk failed")

		return nil, err
	}

	metrics := c.parseDDMMetrics(ctx, ddmData)

	span.SetAttributes(attribute.Int("metrics.count", len(metrics)))

	return &DDMResult{
		SysName: ddmData.sysName,
		Metrics: metrics,
	}, nil
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

// pduToString extracts a string value from an SNMP PDU, returning false for
// unsupported types.
func pduToString(pdu gosnmp.SnmpPDU) (string, bool) {
	//nolint:exhaustive // only OctetString and Integer are expected from TP-Link DDM
	switch pdu.Type {
	case gosnmp.OctetString:
		return string(pdu.Value.([]byte)), true
	case gosnmp.Integer:
		return strconv.Itoa(pdu.Value.(int)), true
	default:
		return "", false
	}
}

// buildOIDDispatch creates a mapping from OID column prefix to the
// corresponding field in ddmWalkData. Used to dispatch PDUs from a single
// root BulkWalk into the correct slices.
func buildOIDDispatch(data *ddmWalkData) map[string]*[]string {
	return map[string]*[]string{
		oidDDMStatusPort:        &data.ports,
		oidDDMStatusTemperature: &data.temps,
		oidDDMStatusVoltage:     &data.voltages,
		oidDDMStatusBiasCurrent: &data.biasCurrents,
		oidDDMStatusTxPower:     &data.txPowers,
		oidDDMStatusRxPower:     &data.rxPowers,
		oidDDMStatusSupported:   &data.ddmSupported,
		oidDDMStatusLossSignal:  &data.lossOfSignal,
		oidDDMStatusTxFault:     &data.txFault,

		oidDDMConfigStatus:   &data.ddmEnabled,
		oidDDMConfigShutdown: &data.shutdownPolicy,
		oidDDMConfigPortLAG:  &data.lagMembership,

		oidDDMRxPowerHighAlarm:   &data.rxPowerHighAlarm,
		oidDDMRxPowerLowAlarm:    &data.rxPowerLowAlarm,
		oidDDMRxPowerHighWarning: &data.rxPowerHighWarning,
		oidDDMRxPowerLowWarning:  &data.rxPowerLowWarning,

		oidDDMVoltageHighAlarm:   &data.voltageHighAlarm,
		oidDDMVoltageLowAlarm:    &data.voltageLowAlarm,
		oidDDMVoltageHighWarning: &data.voltageHighWarning,
		oidDDMVoltageLowWarning:  &data.voltageLowWarning,

		oidDDMBiasCurrentHighAlarm:   &data.biasCurrentHighAlarm,
		oidDDMBiasCurrentLowAlarm:    &data.biasCurrentLowAlarm,
		oidDDMBiasCurrentHighWarning: &data.biasCurrentHighWarning,
		oidDDMBiasCurrentLowWarning:  &data.biasCurrentLowWarning,

		oidDDMTxPowerHighAlarm:   &data.txPowerHighAlarm,
		oidDDMTxPowerLowAlarm:    &data.txPowerLowAlarm,
		oidDDMTxPowerHighWarning: &data.txPowerHighWarning,
		oidDDMTxPowerLowWarning:  &data.txPowerLowWarning,

		oidDDMTemperatureHighAlarm:   &data.tempHighAlarm,
		oidDDMTemperatureLowAlarm:    &data.tempLowAlarm,
		oidDDMTemperatureHighWarning: &data.tempHighWarning,
		oidDDMTemperatureLowWarning:  &data.tempLowWarning,
	}
}

// dispatchPDU routes a single PDU to the correct ddmWalkData field based on
// its OID prefix.
func dispatchPDU(pdu gosnmp.SnmpPDU, dispatch map[string]*[]string) {
	val, ok := pduToString(pdu)
	if !ok {
		slog.Debug("unexpected SNMP type in root walk", "oid", pdu.Name, "type", pdu.Type)

		return
	}

	for prefix, field := range dispatch {
		if strings.HasPrefix(pdu.Name, "."+prefix+".") {
			*field = append(*field, val)

			return
		}
	}
}

func (c *SNMPClient) getSysName(ctx context.Context, client *gosnmp.GoSNMP) string {
	_, span := tracer.Start(ctx, "SNMPClient.getSysName")
	defer span.End()

	client.Context = ctx

	result, err := client.Get([]string{oidSysName})
	if err != nil {
		slog.Debug("failed to get sysName", "error", err)
		span.RecordError(err)

		return ""
	}

	if len(result.Variables) > 0 {
		pdu := result.Variables[0]
		if pdu.Type == gosnmp.OctetString {
			name := string(pdu.Value.([]byte))
			span.SetAttributes(attribute.String("snmp.sysName", name))

			return name
		}
	}

	return ""
}

func (c *SNMPClient) walkAllOIDs(ctx context.Context, client *gosnmp.GoSNMP) (*ddmWalkData, error) {
	ctx, span := tracer.Start(ctx, "SNMPClient.walkAllOIDs")
	defer span.End()

	data := &ddmWalkData{}

	data.sysName = c.getSysName(ctx, client)

	dispatch := buildOIDDispatch(data)
	client.Context = ctx

	var pduCount int

	err := client.BulkWalk(oidDDMRoot, func(pdu gosnmp.SnmpPDU) error {
		pduCount++

		dispatchPDU(pdu, dispatch)

		return nil
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "root walk failed")

		return nil, fmt.Errorf("DDM root walk failed: %w", err)
	}

	span.SetAttributes(
		attribute.Int("snmp.pdu_count", pduCount),
		attribute.Int("snmp.ports", len(data.ports)),
	)

	if len(data.ports) == 0 {
		return nil, fmt.Errorf("no DDM port data found in walk of %s", oidDDMRoot)
	}

	return data, nil
}

//nolint:gocognit,gocyclo,funlen // Parsing many DDM threshold fields
func (c *SNMPClient) parseDDMMetrics(ctx context.Context, data *ddmWalkData) []DDMMetrics {
	_, span := tracer.Start(ctx, "SNMPClient.parseDDMMetrics",
		trace.WithAttributes(attribute.Int("port.count", len(data.ports))),
	)
	defer span.End()

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
