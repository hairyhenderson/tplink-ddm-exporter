package main

import (
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Exporter collects DDM metrics from TP-Link switch
type Exporter struct {
	snmpClient *SNMPClient
	up         prometheus.Gauge
	duration   prometheus.Gauge
	temp       *prometheus.GaugeVec
	voltage    *prometheus.GaugeVec
	biasCurr   *prometheus.GaugeVec
	txPower    *prometheus.GaugeVec
	rxPower    *prometheus.GaugeVec
}

// NewExporter creates a new DDM exporter
func NewExporter(snmpClient *SNMPClient) *Exporter {
	return &Exporter{
		snmpClient: snmpClient,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "tplink_ddm_exporter_up",
			Help: "Was the last scrape successful (1 = yes, 0 = no)",
		}),
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "tplink_ddm_scrape_duration_seconds",
			Help: "Duration of the last scrape in seconds",
		}),
		temp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_temperature_celsius",
				Help: "SFP temperature in Celsius",
			},
			[]string{"port"},
		),
		voltage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_voltage_volts",
				Help: "SFP voltage in volts",
			},
			[]string{"port"},
		),
		biasCurr: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_bias_current_milliamps",
				Help: "SFP bias current in milliamps",
			},
			[]string{"port"},
		),
		txPower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_tx_power_dbm",
				Help: "SFP TX power in dBm",
			},
			[]string{"port"},
		),
		rxPower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tplink_sfp_rx_power_dbm",
				Help: "SFP RX power in dBm",
			},
			[]string{"port"},
		),
	}
}

// Describe implements prometheus.Collector
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.up.Describe(ch)
	e.duration.Describe(ch)
	e.temp.Describe(ch)
	e.voltage.Describe(ch)
	e.biasCurr.Describe(ch)
	e.txPower.Describe(ch)
	e.rxPower.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	defer func() {
		e.duration.Set(time.Since(start).Seconds())
		e.duration.Collect(ch)
	}()

	// Scrape metrics
	metrics, err := e.snmpClient.GetDDMMetrics()
	if err != nil {
		slog.Error("failed to scrape metrics", "error", err)
		e.up.Set(0)
		e.up.Collect(ch)
		return
	}

	e.up.Set(1)
	e.up.Collect(ch)

	// Reset all metrics
	e.temp.Reset()
	e.voltage.Reset()
	e.biasCurr.Reset()
	e.txPower.Reset()
	e.rxPower.Reset()

	// Set new values
	for _, m := range metrics {
		e.temp.WithLabelValues(m.Port).Set(m.Temperature)
		e.voltage.WithLabelValues(m.Port).Set(m.Voltage)
		e.biasCurr.WithLabelValues(m.Port).Set(m.BiasCurrent)
		e.txPower.WithLabelValues(m.Port).Set(m.TxPower)
		e.rxPower.WithLabelValues(m.Port).Set(m.RxPower)
	}

	// Collect all metrics
	e.temp.Collect(ch)
	e.voltage.Collect(ch)
	e.biasCurr.Collect(ch)
	e.txPower.Collect(ch)
	e.rxPower.Collect(ch)
}

