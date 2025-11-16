package tplinkddm

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// SNMPGetter defines the interface for getting DDM metrics via SNMP
type SNMPGetter interface {
	GetDDMMetrics(ctx context.Context) (*DDMResult, error)
}

// Collector collects DDM metrics from TP-Link switch
type Collector struct { //nolint:govet // field grouping by category is clearer than optimal alignment
	snmpClient SNMPGetter
	target     string

	// Current values
	temp     *prometheus.GaugeVec
	voltage  *prometheus.GaugeVec
	biasCurr *prometheus.GaugeVec
	txPower  *prometheus.GaugeVec
	rxPower  *prometheus.GaugeVec

	// Configuration
	ddmEnabled     *prometheus.GaugeVec
	shutdownPolicy *prometheus.GaugeVec
	portLAG        *prometheus.GaugeVec

	// Status flags
	ddmSupported *prometheus.GaugeVec
	lossOfSignal *prometheus.GaugeVec
	txFault      *prometheus.GaugeVec

	// Thresholds with labels: device, target, port, level (high/low), type (alarm/warning)
	tempThreshold        *prometheus.GaugeVec
	voltageThreshold     *prometheus.GaugeVec
	biasCurrentThreshold *prometheus.GaugeVec
	txPowerThreshold     *prometheus.GaugeVec
	rxPowerThreshold     *prometheus.GaugeVec
}

// NewCollector creates a new DDM collector for a given target
//
//nolint:funlen,dupl // Multiple metric definitions required
func NewCollector(snmpClient *SNMPClient, target string) *Collector {
	labels := []string{"device", "target", "port"}
	thresholdLabels := []string{"device", "target", "port", "level", "type"}

	return &Collector{
		snmpClient: snmpClient,
		target:     target,
		// Current values
		temp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_temperature_celsius",
				Help: "SFP temperature in Celsius",
			},
			labels,
		),
		voltage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_voltage_volts",
				Help: "SFP voltage in volts",
			},
			labels,
		),
		biasCurr: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_bias_current_amperes",
				Help: "SFP bias current in amperes",
			},
			labels,
		),
		txPower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_tx_power_dbm",
				Help: "SFP TX power in dBm",
			},
			labels,
		),
		rxPower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_rx_power_dbm",
				Help: "SFP RX power in dBm",
			},
			labels,
		),
		// Configuration
		ddmEnabled: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_ddm_enabled",
				Help: "Whether DDM monitoring is enabled on the port (1 = enabled, 0 = disabled)",
			},
			labels,
		),
		shutdownPolicy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_ddm_shutdown_policy",
				Help: "Port shutdown policy on threshold violation (0 = none, 1 = warning, 2 = alarm)",
			},
			labels,
		),
		portLAG: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_port_lag_member",
				Help: "Port LAG/trunk membership (1 = member, 0 = not member). LAG name in label.",
			},
			[]string{"device", "target", "port", "lag"},
		),
		// Status flags
		ddmSupported: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_ddm_supported",
				Help: "Whether the SFP supports DDM (1 = yes, 0 = no)",
			},
			labels,
		),
		lossOfSignal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_loss_of_signal",
				Help: "SFP Loss of Signal status (1 = signal lost, 0 = ok)",
			},
			labels,
		),
		txFault: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_tx_fault",
				Help: "SFP transmitter fault status (1 = fault, 0 = ok)",
			},
			labels,
		),
		// Thresholds (consolidated with labels)
		tempThreshold: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_temperature_threshold_celsius",
				Help: "SFP temperature threshold in Celsius",
			},
			thresholdLabels,
		),
		voltageThreshold: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_voltage_threshold_volts",
				Help: "SFP voltage threshold in volts",
			},
			thresholdLabels,
		),
		biasCurrentThreshold: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_bias_current_threshold_amperes",
				Help: "SFP bias current threshold in amperes",
			},
			thresholdLabels,
		),
		txPowerThreshold: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_tx_power_threshold_dbm",
				Help: "SFP TX power threshold in dBm",
			},
			thresholdLabels,
		),
		rxPowerThreshold: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_rx_power_threshold_dbm",
				Help: "SFP RX power threshold in dBm",
			},
			thresholdLabels,
		),
	}
}

// Describe implements prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.temp.Describe(ch)
	c.voltage.Describe(ch)
	c.biasCurr.Describe(ch)
	c.txPower.Describe(ch)
	c.rxPower.Describe(ch)
	c.ddmEnabled.Describe(ch)
	c.shutdownPolicy.Describe(ch)
	c.portLAG.Describe(ch)
	c.ddmSupported.Describe(ch)
	c.lossOfSignal.Describe(ch)
	c.txFault.Describe(ch)
	c.tempThreshold.Describe(ch)
	c.voltageThreshold.Describe(ch)
	c.biasCurrentThreshold.Describe(ch)
	c.txPowerThreshold.Describe(ch)
	c.rxPowerThreshold.Describe(ch)
}

// Collect implements prometheus.Collector
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	c.CollectWithContext(ctx, ch)
}

// CollectWithContext collects metrics with a context for cancellation support
//
//nolint:funlen // Multiple metrics to collect and export
func (c *Collector) CollectWithContext(ctx context.Context, ch chan<- prometheus.Metric) {
	result, err := c.snmpClient.GetDDMMetrics(ctx)
	if err != nil {
		slog.Error("failed to scrape metrics", "error", err)

		return
	}

	device := result.SysName

	// Reset all metrics
	c.temp.Reset()
	c.voltage.Reset()
	c.biasCurr.Reset()
	c.txPower.Reset()
	c.rxPower.Reset()
	c.ddmEnabled.Reset()
	c.shutdownPolicy.Reset()
	c.portLAG.Reset()
	c.ddmSupported.Reset()
	c.lossOfSignal.Reset()
	c.txFault.Reset()
	c.tempThreshold.Reset()
	c.voltageThreshold.Reset()
	c.biasCurrentThreshold.Reset()
	c.txPowerThreshold.Reset()
	c.rxPowerThreshold.Reset()

	for _, m := range result.Metrics {
		// Current values
		c.temp.WithLabelValues(device, c.target, m.Port).Set(m.Temperature)
		c.voltage.WithLabelValues(device, c.target, m.Port).Set(m.Voltage)
		c.biasCurr.WithLabelValues(device, c.target, m.Port).Set(m.BiasCurrent / 1000)
		c.txPower.WithLabelValues(device, c.target, m.Port).Set(m.TxPower)
		c.rxPower.WithLabelValues(device, c.target, m.Port).Set(m.RxPower)

		// Configuration
		if m.DDMEnabled {
			c.ddmEnabled.WithLabelValues(device, c.target, m.Port).Set(1)
		} else {
			c.ddmEnabled.WithLabelValues(device, c.target, m.Port).Set(0)
		}

		c.shutdownPolicy.WithLabelValues(device, c.target, m.Port).Set(float64(m.ShutdownPolicy))

		if m.LAGMembership != "" && m.LAGMembership != "N/A" && m.LAGMembership != "---" {
			c.portLAG.WithLabelValues(device, c.target, m.Port, m.LAGMembership).Set(1)
		} else {
			c.portLAG.WithLabelValues(device, c.target, m.Port, "").Set(0)
		}

		// Status flags
		if m.DDMSupported {
			c.ddmSupported.WithLabelValues(device, c.target, m.Port).Set(1)
		} else {
			c.ddmSupported.WithLabelValues(device, c.target, m.Port).Set(0)
		}

		if m.LossOfSignal {
			c.lossOfSignal.WithLabelValues(device, c.target, m.Port).Set(1)
		} else {
			c.lossOfSignal.WithLabelValues(device, c.target, m.Port).Set(0)
		}

		if m.TxFault {
			c.txFault.WithLabelValues(device, c.target, m.Port).Set(1)
		} else {
			c.txFault.WithLabelValues(device, c.target, m.Port).Set(0)
		}

		// Thresholds
		c.tempThreshold.WithLabelValues(device, c.target, m.Port, "high", "alarm").Set(m.TemperatureHighAlarm)
		c.tempThreshold.WithLabelValues(device, c.target, m.Port, "low", "alarm").Set(m.TemperatureLowAlarm)
		c.tempThreshold.WithLabelValues(device, c.target, m.Port, "high", "warning").Set(m.TemperatureHighWarning)
		c.tempThreshold.WithLabelValues(device, c.target, m.Port, "low", "warning").Set(m.TemperatureLowWarning)

		c.voltageThreshold.WithLabelValues(device, c.target, m.Port, "high", "alarm").Set(m.VoltageHighAlarm)
		c.voltageThreshold.WithLabelValues(device, c.target, m.Port, "low", "alarm").Set(m.VoltageLowAlarm)
		c.voltageThreshold.WithLabelValues(device, c.target, m.Port, "high", "warning").Set(m.VoltageHighWarning)
		c.voltageThreshold.WithLabelValues(device, c.target, m.Port, "low", "warning").Set(m.VoltageLowWarning)

		c.biasCurrentThreshold.WithLabelValues(device, c.target, m.Port, "high", "alarm").Set(m.BiasCurrentHighAlarm / 1000)
		c.biasCurrentThreshold.WithLabelValues(device, c.target, m.Port, "low", "alarm").Set(m.BiasCurrentLowAlarm / 1000)
		c.biasCurrentThreshold.WithLabelValues(device, c.target, m.Port, "high", "warning").Set(m.BiasCurrentHighWarning / 1000)
		c.biasCurrentThreshold.WithLabelValues(device, c.target, m.Port, "low", "warning").Set(m.BiasCurrentLowWarning / 1000)

		c.txPowerThreshold.WithLabelValues(device, c.target, m.Port, "high", "alarm").Set(m.TxPowerHighAlarm)
		c.txPowerThreshold.WithLabelValues(device, c.target, m.Port, "low", "alarm").Set(m.TxPowerLowAlarm)
		c.txPowerThreshold.WithLabelValues(device, c.target, m.Port, "high", "warning").Set(m.TxPowerHighWarning)
		c.txPowerThreshold.WithLabelValues(device, c.target, m.Port, "low", "warning").Set(m.TxPowerLowWarning)

		c.rxPowerThreshold.WithLabelValues(device, c.target, m.Port, "high", "alarm").Set(m.RxPowerHighAlarm)
		c.rxPowerThreshold.WithLabelValues(device, c.target, m.Port, "low", "alarm").Set(m.RxPowerLowAlarm)
		c.rxPowerThreshold.WithLabelValues(device, c.target, m.Port, "high", "warning").Set(m.RxPowerHighWarning)
		c.rxPowerThreshold.WithLabelValues(device, c.target, m.Port, "low", "warning").Set(m.RxPowerLowWarning)
	}

	// Collect all metrics
	c.temp.Collect(ch)
	c.voltage.Collect(ch)
	c.biasCurr.Collect(ch)
	c.txPower.Collect(ch)
	c.rxPower.Collect(ch)
	c.ddmEnabled.Collect(ch)
	c.shutdownPolicy.Collect(ch)
	c.portLAG.Collect(ch)
	c.ddmSupported.Collect(ch)
	c.lossOfSignal.Collect(ch)
	c.txFault.Collect(ch)
	c.tempThreshold.Collect(ch)
	c.voltageThreshold.Collect(ch)
	c.biasCurrentThreshold.Collect(ch)
	c.txPowerThreshold.Collect(ch)
	c.rxPowerThreshold.Collect(ch)
}
