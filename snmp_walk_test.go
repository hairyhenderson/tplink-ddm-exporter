package tplinkddm

import (
	"testing"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPduToString(t *testing.T) {
	tests := []struct { //nolint:govet // test struct
		pdu  gosnmp.SnmpPDU
		name string
		want string
		ok   bool
	}{
		{
			name: "OctetString",
			pdu:  gosnmp.SnmpPDU{Type: gosnmp.OctetString, Value: []byte("45.5")},
			want: "45.5",
			ok:   true,
		},
		{
			name: "Integer",
			pdu:  gosnmp.SnmpPDU{Type: gosnmp.Integer, Value: 1},
			want: "1",
			ok:   true,
		},
		{
			name: "unsupported type",
			pdu:  gosnmp.SnmpPDU{Type: gosnmp.Counter32, Value: uint(42)},
			ok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := pduToString(tt.pdu)
			assert.Equal(t, tt.ok, ok)

			if ok {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestBuildOIDDispatch(t *testing.T) {
	data := &ddmWalkData{}
	dispatch := buildOIDDispatch(data)

	assert.Contains(t, dispatch, oidDDMStatusPort)
	assert.Contains(t, dispatch, oidDDMStatusTemperature)
	assert.Contains(t, dispatch, oidDDMConfigStatus)
	assert.Contains(t, dispatch, oidDDMTemperatureHighAlarm)
	assert.Contains(t, dispatch, oidDDMRxPowerLowWarning)

	assert.Len(t, dispatch, 32)
}

func TestDispatchPDU(t *testing.T) {
	data := &ddmWalkData{}
	dispatch := buildOIDDispatch(data)

	pdus := []gosnmp.SnmpPDU{
		{Name: "." + oidDDMStatusPort + ".49153", Type: gosnmp.OctetString, Value: []byte("1/0/1")},
		{Name: "." + oidDDMStatusPort + ".49154", Type: gosnmp.OctetString, Value: []byte("1/0/2")},
		{Name: "." + oidDDMStatusTemperature + ".49153", Type: gosnmp.OctetString, Value: []byte("45.5")},
		{Name: "." + oidDDMStatusTemperature + ".49154", Type: gosnmp.OctetString, Value: []byte("46.0")},
		{Name: "." + oidDDMStatusVoltage + ".49153", Type: gosnmp.OctetString, Value: []byte("3.30")},
		{Name: "." + oidDDMStatusVoltage + ".49154", Type: gosnmp.OctetString, Value: []byte("3.29")},
		{Name: "." + oidDDMStatusBiasCurrent + ".49153", Type: gosnmp.OctetString, Value: []byte("6.0")},
		{Name: "." + oidDDMStatusTxPower + ".49153", Type: gosnmp.OctetString, Value: []byte("0.5")},
		{Name: "." + oidDDMStatusRxPower + ".49153", Type: gosnmp.OctetString, Value: []byte("0.4")},
		{Name: "." + oidDDMConfigStatus + ".49153", Type: gosnmp.Integer, Value: 1},
		{Name: "." + oidDDMConfigShutdown + ".49153", Type: gosnmp.Integer, Value: 2},
		{Name: "." + oidDDMConfigPortLAG + ".49153", Type: gosnmp.OctetString, Value: []byte("N/A")},
		{Name: "." + oidDDMStatusSupported + ".49153", Type: gosnmp.Integer, Value: 1},
		{Name: "." + oidDDMStatusLossSignal + ".49153", Type: gosnmp.Integer, Value: 0},
		{Name: "." + oidDDMStatusTxFault + ".49153", Type: gosnmp.Integer, Value: 0},
		{Name: "." + oidDDMTemperatureHighAlarm + ".49153", Type: gosnmp.OctetString, Value: []byte("80.0")},
		{Name: "." + oidDDMRxPowerLowWarning + ".49153", Type: gosnmp.OctetString, Value: []byte("-18.0")},
	}

	for _, pdu := range pdus {
		dispatchPDU(pdu, dispatch)
	}

	require.Equal(t, []string{"1/0/1", "1/0/2"}, data.ports)
	require.Equal(t, []string{"45.5", "46.0"}, data.temps)
	require.Equal(t, []string{"3.30", "3.29"}, data.voltages)
	require.Equal(t, []string{"6.0"}, data.biasCurrents)
	require.Equal(t, []string{"0.5"}, data.txPowers)
	require.Equal(t, []string{"0.4"}, data.rxPowers)
	require.Equal(t, []string{"1"}, data.ddmEnabled)
	require.Equal(t, []string{"2"}, data.shutdownPolicy)
	require.Equal(t, []string{"N/A"}, data.lagMembership)
	require.Equal(t, []string{"1"}, data.ddmSupported)
	require.Equal(t, []string{"0"}, data.lossOfSignal)
	require.Equal(t, []string{"0"}, data.txFault)
	require.Equal(t, []string{"80.0"}, data.tempHighAlarm)
	require.Equal(t, []string{"-18.0"}, data.rxPowerLowWarning)
}

func TestDispatchPDU_IgnoresUnknownOIDs(t *testing.T) {
	data := &ddmWalkData{}
	dispatch := buildOIDDispatch(data)

	dispatchPDU(gosnmp.SnmpPDU{
		Name:  ".1.3.6.1.4.1.11863.6.96.1.99.1.1.1.49153",
		Type:  gosnmp.OctetString,
		Value: []byte("unknown"),
	}, dispatch)

	assert.Empty(t, data.ports)
	assert.Empty(t, data.temps)
}

func TestDispatchPDU_IgnoresUnsupportedTypes(t *testing.T) {
	data := &ddmWalkData{}
	dispatch := buildOIDDispatch(data)

	dispatchPDU(gosnmp.SnmpPDU{
		Name:  "." + oidDDMStatusPort + ".49153",
		Type:  gosnmp.Counter32,
		Value: uint(42),
	}, dispatch)

	assert.Empty(t, data.ports)
}
