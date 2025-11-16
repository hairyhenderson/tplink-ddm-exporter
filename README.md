# TP-Link DDM Exporter

Prometheus exporter for TP-Link switch SFP+ Digital Diagnostic Monitoring (DDM) metrics.

## Problem

TP-Link switches expose DDM metrics (temperature, voltage, bias current, TX/RX power) via SNMP, but encode them as DisplayString (OCTET STRING) instead of proper numeric types. This makes them incompatible with standard SNMP exporters like `snmp_exporter`.

## Solution

This exporter queries the TP-Link DDM SNMP OIDs directly, parses the string values, and exposes them as proper Prometheus metrics.

## Metrics

### Current Values

```
tplink_sfp_temperature_celsius{device="...",target="...",port="N"} - SFP temperature in Celsius
tplink_sfp_voltage_volts{device="...",target="...",port="N"} - SFP voltage in volts
tplink_sfp_bias_current_amperes{device="...",target="...",port="N"} - SFP bias current in amperes
tplink_sfp_tx_power_dbm{device="...",target="...",port="N"} - SFP TX power in dBm
tplink_sfp_rx_power_dbm{device="...",target="...",port="N"} - SFP RX power in dBm
```

### Configuration

```
tplink_ddm_enabled{device="...",target="...",port="N"} - Whether DDM monitoring is enabled on the port (1 = enabled, 0 = disabled)
tplink_ddm_shutdown_policy{device="...",target="...",port="N"} - Port shutdown policy on threshold violation (0 = none, 1 = warning, 2 = alarm)
tplink_port_lag_member{device="...",target="...",port="N",lag="name"} - Port LAG/trunk membership (1 = member, 0 = not member)
```

These are configuration settings that control DDM behavior and port membership.

### Status Flags

```
tplink_sfp_ddm_supported{device="...",target="...",port="N"} - Whether the SFP supports DDM (1 = yes, 0 = no)
tplink_sfp_loss_of_signal{device="...",target="...",port="N"} - Loss of Signal status (1 = signal lost, 0 = ok)
tplink_sfp_tx_fault{device="...",target="...",port="N"} - Transmitter fault status (1 = fault, 0 = ok)
```

These status flags come from the SFP module's internal diagnostics and indicate real-time operational issues.

### Thresholds

These metrics are static values burned into the SFP module's EEPROM at manufacturing time. They define the safe operating ranges for the transceiver. All thresholds use labels to distinguish between high/low thresholds and alarm/warning types:

```
tplink_sfp_temperature_threshold_celsius{device="...",target="...",port="N", level="high|low", type="alarm|warning"}
tplink_sfp_voltage_threshold_volts{device="...",target="...",port="N", level="high|low", type="alarm|warning"}
tplink_sfp_bias_current_threshold_amperes{device="...",target="...",port="N", level="high|low", type="alarm|warning"}
tplink_sfp_tx_power_threshold_dbm{device="...",target="...",port="N", level="high|low", type="alarm|warning"}
tplink_sfp_rx_power_threshold_dbm{device="...",target="...",port="N", level="high|low", type="alarm|warning"}
```

Example queries:
- High alarm threshold for temperature on port 1: `tplink_sfp_temperature_threshold_celsius{port="1", level="high", type="alarm"}`
- All low warning thresholds: `{__name__=~"tplink_sfp_.*_threshold_.*", level="low", type="warning"}`
- Compare current temperature to high alarm: `tplink_sfp_temperature_celsius > on(port) tplink_sfp_temperature_threshold_celsius{level="high", type="alarm"}`
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

