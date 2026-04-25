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

func TestParseDDMMetrics_OptionalFields(t *testing.T) {
	client := &SNMPClient{}

	t.Run("with all optional fields", func(t *testing.T) {
		data := &ddmWalkData{
			sysName:                "Test Switch",
			ports:                  []string{"1/0/1"},
			temps:                  []string{"45.5"},
			voltages:               []string{"3.30"},
			biasCurrents:           []string{"6.0"},
			txPowers:               []string{"0.5"},
			rxPowers:               []string{"0.4"},
			ddmEnabled:             []string{"1"},
			shutdownPolicy:         []string{"2"},
			lagMembership:          []string{"Trunk1"},
			ddmSupported:           []string{"1"},
			lossOfSignal:           []string{"0"},
			txFault:                []string{"0"},
			tempHighAlarm:          []string{"80.0"},
			tempLowAlarm:           []string{"-10.0"},
			tempHighWarning:        []string{"70.0"},
			tempLowWarning:         []string{"0.0"},
			voltageHighAlarm:       []string{"3.6"},
			voltageLowAlarm:        []string{"2.9"},
			voltageHighWarning:     []string{"3.5"},
			voltageLowWarning:      []string{"3.0"},
			biasCurrentHighAlarm:   []string{"85.0"},
			biasCurrentLowAlarm:    []string{"1.0"},
			biasCurrentHighWarning: []string{"70.0"},
			biasCurrentLowWarning:  []string{"2.0"},
			txPowerHighAlarm:       []string{"1.0"},
			txPowerLowAlarm:        []string{"-5.0"},
			txPowerHighWarning:     []string{"0.5"},
			txPowerLowWarning:      []string{"-4.0"},
			rxPowerHighAlarm:       []string{"1.0"},
			rxPowerLowAlarm:        []string{"-20.0"},
			rxPowerHighWarning:     []string{"0.5"},
			rxPowerLowWarning:      []string{"-18.0"},
		}

		metrics := client.parseDDMMetrics(data)
		require.Len(t, metrics, 1)

		m := metrics[0]
		assert.True(t, m.DDMEnabled)
		assert.Equal(t, 2, m.ShutdownPolicy)
		assert.Equal(t, "Trunk1", m.LAGMembership)
		assert.True(t, m.DDMSupported)
		assert.False(t, m.LossOfSignal)
		assert.False(t, m.TxFault)
		assert.InDelta(t, 80.0, m.TemperatureHighAlarm, 0.01)
		assert.InDelta(t, -10.0, m.TemperatureLowAlarm, 0.01)
		assert.InDelta(t, 3.6, m.VoltageHighAlarm, 0.01)
		assert.InDelta(t, 85.0, m.BiasCurrentHighAlarm, 0.01)
		assert.InDelta(t, -5.0, m.TxPowerLowAlarm, 0.01)
		assert.InDelta(t, -20.0, m.RxPowerLowAlarm, 0.01)
	})

	t.Run("with nil optional fields", func(t *testing.T) {
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
		assert.InDelta(t, 45.5, m.Temperature, 0.01)
		assert.False(t, m.DDMEnabled)
		assert.Equal(t, 0, m.ShutdownPolicy)
		assert.Empty(t, m.LAGMembership)
		assert.False(t, m.DDMSupported)
		assert.InDelta(t, 0, m.TemperatureHighAlarm, 0.01)
	})
}
