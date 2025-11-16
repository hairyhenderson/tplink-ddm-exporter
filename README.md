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

Command-line flags:
- `-target` - SNMP target IP address (default: `192.168.2.96`)
- `-community` - SNMP community string (default: `public`)
- `-addr` - Listen address (default: `:9116`)
- `-log-level` - Log level: debug, info, warn, error (default: `info`)
- `-version` - Show version and exit

OpenTelemetry tracing can be configured via standard OTEL environment variables:
- `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` or `OTEL_EXPORTER_OTLP_ENDPOINT` - OTLP endpoint URL
- `OTEL_EXPORTER_OTLP_TRACES_INSECURE` or `OTEL_EXPORTER_OTLP_INSECURE` - Set to `true` for non-TLS endpoints
- `OTEL_SDK_DISABLED` - Set to `true` to disable tracing

## Endpoints

- `/metrics` - Exporter self-metrics (up, duration, go/process metrics)
- `/scrape` - Scraped device metrics (SFP temperature, voltage, power, etc.)
- `/` - HTML status page

## Usage

### Binary
```bash
./tplink-ddm-exporter -target 192.168.2.96 -community public -addr :9116
```

### Docker
```bash
docker run -d \
  --name tplink-ddm-exporter \
  --network host \
  ghcr.io/hairyhenderson/tplink-ddm-exporter:latest \
  -target 192.168.2.96 -community public -addr :9116
```

## Development

```bash
# Run tests
go test -v ./...

# Build
go build -o tplink-ddm-exporter ./cmd/tplink-ddm-exporter

# Run
./tplink-ddm-exporter
```

