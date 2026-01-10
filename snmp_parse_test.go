package tplinkddm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDDMMetrics(t *testing.T) {
	client := &SNMPClient{}

	tests := []struct {
		data *ddmWalkData
		name string
		want int // expected number of metrics
	}{
		{
			name: "valid data",
			data: &ddmWalkData{
				sysName:      "Test Switch",
				ports:        []string{"1/0/1", "1/0/2"},
				temps:        []string{"45.5", "46.0"},
				voltages:     []string{"3.30", "3.29"},
				biasCurrents: []string{"6.0", "5.8"},
				txPowers:     []string{"0.5", "0.6"},
				rxPowers:     []string{"0.4", "0.45"},
			},
			want: 2,
		},
		{
			name: "mismatched lengths",
			data: &ddmWalkData{
				sysName:      "Test Switch",
				ports:        []string{"1/0/1", "1/0/2", "1/0/3"},
				temps:        []string{"45.5", "46.0"}, // shorter
				voltages:     []string{"3.30"},         // even shorter
				biasCurrents: []string{"6.0", "5.8", "5.9"},
				txPowers:     []string{"0.5"},
				rxPowers:     []string{"0.4", "0.45", "0.43"},
			},
			want: 3, // should handle all ports
		},
		{
			name: "invalid port format",
			data: &ddmWalkData{
				sysName:      "Test Switch",
				ports:        []string{"invalid", "1/0/2"},
				temps:        []string{"45.5", "46.0"},
				voltages:     []string{"3.30", "3.29"},
				biasCurrents: []string{"6.0", "5.8"},
				txPowers:     []string{"0.5", "0.6"},
				rxPowers:     []string{"0.4", "0.45"},
			},
			want: 1, // invalid port should be skipped
		},
		{
			name: "invalid float values",
			data: &ddmWalkData{
				sysName:      "Test Switch",
				ports:        []string{"1/0/1"},
				temps:        []string{"invalid"},
				voltages:     []string{"not-a-number"},
				biasCurrents: []string{"bad"},
				txPowers:     []string{"wrong"},
				rxPowers:     []string{"nope"},
			},
			want: 1, // should still create metric with zero values
		},
		{
			name: "empty data",
			data: &ddmWalkData{
				sysName:      "Test Switch",
				ports:        []string{},
				temps:        []string{},
				voltages:     []string{},
				biasCurrents: []string{},
				txPowers:     []string{},
				rxPowers:     []string{},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.parseDDMMetrics(tt.data)
			assert.Len(t, got, tt.want)

			// Verify structure of returned metrics
			for _, m := range got {
				assert.NotEmpty(t, m.Port)
				// Temperature, Voltage, etc. can be zero for invalid inputs, which is fine
			}
		})
	}
}

func TestParseDDMMetrics_Values(t *testing.T) {
	client := &SNMPClient{}

	data := &ddmWalkData{
		sysName:      "Test Switch",
		ports:        []string{"1/0/1"},
		temps:        []string{"45.5"},
		voltages:     []string{"3.30"},
		biasCurrents: []string{"6.0"},
		txPowers:     []string{"0.5"},
		rxPowers:     []string{"0.4"},
	}

	metrics := client.parseDDMMetrics(data)

	require.Len(t, metrics, 1)

	m := metrics[0]

	assert.Equal(t, "1", m.Port)
	assert.InDelta(t, 45.5, m.Temperature, 0.01)
	assert.InDelta(t, 3.30, m.Voltage, 0.01)
	assert.InDelta(t, 6.0, m.BiasCurrent, 0.01)
	assert.InDelta(t, 0.5, m.TxPower, 0.01)
	assert.InDelta(t, 0.4, m.RxPower, 0.01)
}
