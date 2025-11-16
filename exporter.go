package tplinkddm

import (
	"context"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Exporter collects DDM metrics from TP-Link switch
type Exporter struct {
	snmpClient *SNMPClient
	Up         prometheus.Gauge
	Duration   prometheus.Gauge
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
		Up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "tplink_ddm_exporter_up",
			Help: "Was the last scrape successful (1 = yes, 0 = no)",
		}),
		Duration: prometheus.NewGauge(prometheus.GaugeOpts{
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
				Name: "tplink_sfp_bias_current_amperes",
				Help: "SFP bias current in amperes",
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
	e.Up.Describe(ch)
	e.Duration.Describe(ch)
	e.temp.Describe(ch)
	e.voltage.Describe(ch)
	e.biasCurr.Describe(ch)
	e.txPower.Describe(ch)
	e.rxPower.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	e.CollectWithContext(ctx, ch)
}

// CollectWithContext collects metrics with a context for cancellation support
func (e *Exporter) CollectWithContext(ctx context.Context, ch chan<- prometheus.Metric) {
	start := time.Now()

	defer func() {
		e.Duration.Set(time.Since(start).Seconds())
		e.Duration.Collect(ch)
	}()

	// Scrape metrics
	metrics, err := e.snmpClient.GetDDMMetrics(ctx)
	if err != nil {
		slog.Error("failed to scrape metrics", "error", err)
		e.Up.Set(0)
		e.Up.Collect(ch)

		return
	}

	e.Up.Set(1)
	e.Up.Collect(ch)

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
		e.biasCurr.WithLabelValues(m.Port).Set(m.BiasCurrent / 1000) // Convert mA to A
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
