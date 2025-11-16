package tplinkddm

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

// mockSNMPClient implements a mock SNMP client for testing
type mockSNMPClient struct {
	result *DDMResult
	err    error
}

func (m *mockSNMPClient) GetDDMMetrics(_ context.Context) (*DDMResult, error) {
	return m.result, m.err
}

func TestCollector_Collect(t *testing.T) {
	tests := []struct {
		name       string
		mockResult *DDMResult
		mockErr    error
		target     string
		wantCount  int
	}{
		{
			name: "successful collection",
			mockResult: &DDMResult{
				SysName: "Test Switch",
				Metrics: []DDMMetrics{
					{
						Port:        "1",
						Temperature: 45.5,
						Voltage:     3.3,
						BiasCurrent: 6.0,
						TxPower:     0.5,
						RxPower:     0.4,
					},
					{
						Port:        "2",
						Temperature: 46.0,
						Voltage:     3.29,
						BiasCurrent: 5.8,
						TxPower:     0.6,
						RxPower:     0.45,
					},
				},
			},
			target: "192.168.1.1",
			// 5 current + 3 config + 3 status + 5*4 thresholds = 31 per port, * 2 ports = 62
			wantCount: 62,
		},
		{
			name: "empty sysName",
			mockResult: &DDMResult{
				SysName: "",
				Metrics: []DDMMetrics{
					{
						Port:        "1",
						Temperature: 45.5,
						Voltage:     3.3,
						BiasCurrent: 6.0,
						TxPower:     0.5,
						RxPower:     0.4,
					},
				},
			},
			target:    "192.168.1.2",
			wantCount: 31, // 31 metrics * 1 port
		},
		{
			name:      "SNMP error",
			mockErr:   context.DeadlineExceeded,
			target:    "192.168.1.3",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSNMPClient{
				result: tt.mockResult,
				err:    tt.mockErr,
			}

			collector := newTestCollector(mockClient, tt.target)

			registry := prometheus.NewRegistry()
			registry.MustRegister(collector)

			metricCount := testutil.CollectAndCount(collector)
			assert.Equal(t, tt.wantCount, metricCount)
		})
	}
}

func TestCollector_Describe(t *testing.T) {
	collector := NewCollector(&SNMPClient{}, "192.168.1.1")

	ch := make(chan *prometheus.Desc, 20)

	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	// 5 current + 3 config + 3 status + 5 thresholds = 16
	assert.Equal(t, 16, count)
}

//nolint:dupl // test helper intentionally mirrors NewCollector with test-specific metric names
func newTestCollector(mock SNMPGetter, target string) *Collector {
	labels := []string{"device", "target", "port"}
	thresholdLabels := []string{"device", "target", "port", "level", "type"}

	return &Collector{
		snmpClient: mock,
		target:     target,
		temp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "test_temp", Help: "h"},
			labels,
		),
		voltage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "test_voltage", Help: "h"},
			labels,
		),
		biasCurr: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "test_bias", Help: "h"},
			labels,
		),
		txPower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "test_tx", Help: "h"},
			labels,
		),
		rxPower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "test_rx", Help: "h"},
			labels,
		),
		ddmEnabled: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "test_ddm_enabled", Help: "h"},
			labels,
		),
		shutdownPolicy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "test_shutdown_policy", Help: "h"},
			labels,
		),
		portLAG: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "test_lag", Help: "h"},
			[]string{"device", "target", "port", "lag"},
		),
		ddmSupported: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "test_ddm_supported", Help: "h"},
			labels,
		),
		lossOfSignal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "test_los", Help: "h"},
			labels,
		),
		txFault: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "test_tx_fault", Help: "h"},
			labels,
		),
		tempThreshold: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "test_temp_thresh", Help: "h"},
			thresholdLabels,
		),
		voltageThreshold: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "test_voltage_thresh", Help: "h"},
			thresholdLabels,
		),
		biasCurrentThreshold: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "test_bias_thresh", Help: "h"},
			thresholdLabels,
		),
		txPowerThreshold: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "test_tx_thresh", Help: "h"},
			thresholdLabels,
		),
		rxPowerThreshold: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "test_rx_thresh", Help: "h"},
			thresholdLabels,
		),
	}
}
