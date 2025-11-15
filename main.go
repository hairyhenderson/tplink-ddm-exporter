package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Configuration from environment
	target := getEnv("SNMP_TARGET", "192.168.2.96")
	community := getEnv("SNMP_COMMUNITY", "public")
	listenAddr := getEnv("LISTEN_ADDR", ":9116")

	slog.Info("starting TP-Link DDM exporter",
		"target", target,
		"listen_addr", listenAddr)

	// Create SNMP client and exporter
	snmpClient := NewSNMPClient(target, community)
	exporter := NewExporter(snmpClient)

	// Register exporter
	prometheus.MustRegister(exporter)

	// HTTP handler
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html>
<head><title>TP-Link DDM Exporter</title></head>
<body>
<h1>TP-Link DDM Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
</body>
</html>`))
	})

	slog.Info("listening", "addr", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

