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
		wantCount  int // expected number of metrics
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
			target:    "192.168.1.1",
			wantCount: 10, // 5 metrics * 2 ports
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
			wantCount: 5, // 5 metrics * 1 port
		},
		{
			name:      "SNMP error",
			mockErr:   context.DeadlineExceeded,
			target:    "192.168.1.3",
			wantCount: 0, // no metrics on error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock SNMP client
			mockClient := &mockSNMPClient{
				result: tt.mockResult,
				err:    tt.mockErr,
			}

			// Create collector with mock
			collector := &Collector{
				snmpClient: mockClient,
				target:     tt.target,
				temp: prometheus.NewGaugeVec(
					prometheus.GaugeOpts{Name: "test_temp", Help: "test temperature"},
					[]string{"device", "target", "port"},
				),
				voltage: prometheus.NewGaugeVec(
					prometheus.GaugeOpts{Name: "test_voltage", Help: "test voltage"},
					[]string{"device", "target", "port"},
				),
				biasCurr: prometheus.NewGaugeVec(
					prometheus.GaugeOpts{Name: "test_bias", Help: "test bias current"},
					[]string{"device", "target", "port"},
				),
				txPower: prometheus.NewGaugeVec(
					prometheus.GaugeOpts{Name: "test_tx", Help: "test TX power"},
					[]string{"device", "target", "port"},
				),
				rxPower: prometheus.NewGaugeVec(
					prometheus.GaugeOpts{Name: "test_rx", Help: "test RX power"},
					[]string{"device", "target", "port"},
				),
			}

			// Create a registry and register the collector
			registry := prometheus.NewRegistry()
			registry.MustRegister(collector)

			// Collect metrics
			metricCount := testutil.CollectAndCount(collector)

			// Verify metric count
			assert.Equal(t, tt.wantCount, metricCount)
		})
	}
}

func TestCollector_Describe(t *testing.T) {
	collector := NewCollector(&SNMPClient{}, "192.168.1.1")

	// Collect descriptions
	ch := make(chan *prometheus.Desc, 10)

	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	// Count descriptions
	count := 0
	for range ch {
		count++
	}

	// Should have 5 metric descriptions (temp, voltage, bias, tx, rx)
	assert.Equal(t, 5, count)
}
