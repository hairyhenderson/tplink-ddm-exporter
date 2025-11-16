# TP-Link DDM Exporter

Prometheus exporter for TP-Link switch SFP+ Digital Diagnostic Monitoring (DDM) metrics.

## Problem

TP-Link switches expose DDM metrics (temperature, voltage, bias current, TX/RX power) via SNMP, but encode them as DisplayString (OCTET STRING) instead of proper numeric types. This makes them incompatible with standard SNMP exporters like `snmp_exporter`.

## Solution

This exporter queries the TP-Link DDM SNMP OIDs directly, parses the string values, and exposes them as proper Prometheus metrics.

## Metrics

```
tplink_sfp_temperature_celsius{port="N"} - SFP temperature in Celsius
tplink_sfp_voltage_volts{port="N"} - SFP voltage in volts
tplink_sfp_bias_current_amperes{port="N"} - SFP bias current in amperes
tplink_sfp_tx_power_dbm{port="N"} - SFP TX power in dBm
tplink_sfp_rx_power_dbm{port="N"} - SFP RX power in dBm
tplink_ddm_exporter_up - Exporter health (1 = up, 0 = down)
tplink_ddm_scrape_duration_seconds - Time taken to scrape SNMP metrics
```

## Configuration

Environment variables:
- `SNMP_TARGET` - Switch IP address (default: `192.168.1.1`)
- `SNMP_COMMUNITY` - SNMP community string (default: `public`)
- `LISTEN_ADDR` - Exporter listen address (default: `:9116`)

## Usage

### Binary
```bash
export SNMP_TARGET=192.168.1.1
export SNMP_COMMUNITY=public
./tplink-ddm-exporter
```

### Docker
```bash
docker run -d \
  --name tplink-ddm-exporter \
  --network host \
  -e SNMP_TARGET=192.168.1.1 \
  -e SNMP_COMMUNITY=public \
  ghcr.io/hairyhenderson/tplink-ddm-exporter:latest
```

## Development

```bash
# Run tests
go test -v ./...

# Build
go build -o tplink-ddm-exporter .

# Run
./tplink-ddm-exporter
```

