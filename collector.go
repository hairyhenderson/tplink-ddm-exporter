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
type Collector struct {
	snmpClient SNMPGetter
	temp       *prometheus.GaugeVec
	voltage    *prometheus.GaugeVec
	biasCurr   *prometheus.GaugeVec
	txPower    *prometheus.GaugeVec
	rxPower    *prometheus.GaugeVec
	target     string
}

// NewCollector creates a new DDM collector for a given target
func NewCollector(snmpClient *SNMPClient, target string) *Collector {
	return &Collector{
		snmpClient: snmpClient,
		target:     target,
		temp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_temperature_celsius",
				Help: "SFP temperature in Celsius",
			},
			[]string{"device", "target", "port"},
		),
		voltage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_voltage_volts",
				Help: "SFP voltage in volts",
			},
			[]string{"device", "target", "port"},
		),
		biasCurr: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_bias_current_amperes",
				Help: "SFP bias current in amperes",
			},
			[]string{"device", "target", "port"},
		),
		txPower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_tx_power_dbm",
				Help: "SFP TX power in dBm",
			},
			[]string{"device", "target", "port"},
		),
		rxPower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_rx_power_dbm",
				Help: "SFP RX power in dBm",
			},
			[]string{"device", "target", "port"},
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
}

// Collect implements prometheus.Collector
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	c.CollectWithContext(ctx, ch)
}

// CollectWithContext collects metrics with a context for cancellation support
func (c *Collector) CollectWithContext(ctx context.Context, ch chan<- prometheus.Metric) {
	// Scrape metrics
	result, err := c.snmpClient.GetDDMMetrics(ctx)
	if err != nil {
		slog.Error("failed to scrape metrics", "error", err)

		return
	}

	// Use sysName from result for device label
	device := result.SysName

	// Reset all metrics
	c.temp.Reset()
	c.voltage.Reset()
	c.biasCurr.Reset()
	c.txPower.Reset()
	c.rxPower.Reset()

	// Set new values
	for _, m := range result.Metrics {
		c.temp.WithLabelValues(device, c.target, m.Port).Set(m.Temperature)
		c.voltage.WithLabelValues(device, c.target, m.Port).Set(m.Voltage)
		c.biasCurr.WithLabelValues(device, c.target, m.Port).Set(m.BiasCurrent / 1000) // Convert mA to A
		c.txPower.WithLabelValues(device, c.target, m.Port).Set(m.TxPower)
		c.rxPower.WithLabelValues(device, c.target, m.Port).Set(m.RxPower)
	}

	// Collect all metrics
	c.temp.Collect(ch)
	c.voltage.Collect(ch)
	c.biasCurr.Collect(ch)
	c.txPower.Collect(ch)
	c.rxPower.Collect(ch)
}
