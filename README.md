# TP-Link DDM Exporter

Prometheus exporter for TP-Link switch SFP+ Digital Diagnostic Monitoring (DDM) metrics.

## Problem

TP-Link switches expose DDM metrics (temperature, voltage, bias current, TX/RX power) via SNMP, but encode them as DisplayString (OCTET STRING) instead of proper numeric types. This makes them incompatible with standard SNMP exporters like `snmp_exporter`.

## Solution

This exporter queries the TP-Link DDM SNMP OIDs directly, parses the string values, and exposes them as proper Prometheus metrics.

## Metrics

```
tplink_sfp_temperature_celsius{device="...",target="...",port="N"} - SFP temperature in Celsius
tplink_sfp_voltage_volts{device="...",target="...",port="N"} - SFP voltage in volts
tplink_sfp_bias_current_amperes{device="...",target="...",port="N"} - SFP bias current in amperes
tplink_sfp_tx_power_dbm{device="...",target="...",port="N"} - SFP TX power in dBm
tplink_sfp_rx_power_dbm{device="...",target="...",port="N"} - SFP RX power in dBm
```

All SFP metrics include:
- `device` - Device name (auto-detected via SNMP sysName)
- `target` - SNMP target IP address
- `port` - SFP port number

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

- `/metrics` - Exporter self-metrics (Go runtime and process metrics)
- `/scrape` - Device metrics (SFP temperature, voltage, power, etc.)
  - Query parameters:
    - `target` - SNMP target IP address (defaults to configured target)
    - `community` - SNMP community string (defaults to configured community)
- `/` - HTML status page

## Usage

### Binary
```bash
# Start the exporter with a default target
./tplink-ddm-exporter -target 192.168.2.96 -community public -addr :9116

# Scrape the default target
curl http://localhost:9116/scrape

# Scrape a specific device (device name auto-detected from SNMP)
curl 'http://localhost:9116/scrape?target=192.168.1.100'

# Scrape multiple devices (configure in Prometheus)
# See Prometheus configuration example below
```

### Docker
```bash
docker run -d \
  --name tplink-ddm-exporter \
  --network host \
  ghcr.io/hairyhenderson/tplink-ddm-exporter:latest \
  -target 192.168.2.96 -community public -addr :9116
```

### Prometheus Configuration

To scrape multiple devices, configure Prometheus with static targets:

```yaml
scrape_configs:
  - job_name: 'tplink-ddm'
    static_configs:
      - targets:
        - 192.168.1.100  # switch-core
        - 192.168.1.101  # switch-edge
        - 192.168.1.102  # switch-lab
    metrics_path: /scrape
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9116  # Address of the exporter
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

