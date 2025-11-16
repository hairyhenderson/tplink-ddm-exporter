package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tplinkddm "github.com/hairyhenderson/tplink-ddm-exporter"
	"github.com/hairyhenderson/tplink-ddm-exporter/internal/traceutil"
	"github.com/hairyhenderson/tplink-ddm-exporter/internal/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type config struct {
	Target      string
	Community   string
	ListenAddr  string
	LogLevel    string
	showVersion bool
}

func main() {
	cfg := &config{}
	if err := parseFlags(flag.CommandLine, cfg); err != nil {
		slog.Error("parseFlags", "err", err)
		os.Exit(1)
	}

	if cfg.showVersion {
		fmt.Printf("tplink-ddm-exporter %s (commit: %s)\n", version.Version, version.GitCommit)
		os.Exit(0)
	}

	ctx := context.Background()
	if err := run(ctx, cfg); err != nil {
		slog.ErrorContext(ctx, "exiting with error", "err", err)
		os.Exit(1)
	}
}

func parseFlags(fs *flag.FlagSet, cfg *config) error {
	fs.StringVar(&cfg.Target, "target", "192.168.2.96", "SNMP target IP address")
	fs.StringVar(&cfg.Community, "community", "public", "SNMP community string")
	fs.StringVar(&cfg.ListenAddr, "addr", ":9116", "Listen address")
	fs.StringVar(&cfg.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	fs.BoolVar(&cfg.showVersion, "version", false, "Show version and exit")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	return nil
}

func run(ctx context.Context, cfg *config) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	logger := setupLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	closer, err := traceutil.Init(ctx)
	if err != nil {
		return fmt.Errorf("setupTracing: %w", err)
	}

	if closer != nil {
		defer func(ctx context.Context) {
			// Use WithoutCancel to avoid cancellation affecting shutdown, but derive from parent context
			shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer cancel()

			if shutdownErr := closer(shutdownCtx); shutdownErr != nil {
				logger.Error("failed to shutdown tracer", "error", shutdownErr)
			}
		}(ctx)
	}

	logger.InfoContext(ctx, "starting TP-Link DDM exporter",
		"target", cfg.Target,
		"listen_addr", cfg.ListenAddr)

	// Create SNMP client and exporter
	snmpClient := tplinkddm.NewSNMPClient(cfg.Target, cfg.Community)
	exporter := tplinkddm.NewExporter(snmpClient)

	exporterRegistry, scrapeRegistry := setupRegistries(exporter)
	srv := setupServer(ctx, exporterRegistry, scrapeRegistry)

	return serve(ctx, logger, srv, cfg.ListenAddr, stop)
}

func serve(ctx context.Context, logger *slog.Logger, srv *http.Server, listenAddr string, stop func()) error {
	lc := &net.ListenConfig{}

	ln, err := lc.Listen(ctx, "tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}

	logger.InfoContext(ctx, "listening", slog.String("addr", listenAddr))

	go func() {
		serveErr := srv.Serve(ln)
		if serveErr != nil && serveErr != http.ErrServerClosed {
			logger.ErrorContext(ctx, "server terminated with error", slog.Any("err", serveErr))
		}

		stop()
	}()

	<-ctx.Done()

	logger.InfoContext(ctx, "shutting down gracefully")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer shutdownCancel()

	if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
		return fmt.Errorf("server shutdown: %w", shutdownErr)
	}

	return nil
}

func setupRegistries(exporter *tplinkddm.Exporter) (*prometheus.Registry, *prometheus.Registry) {
	// Create registry for exporter self-metrics only (up, duration)
	exporterRegistry := prometheus.NewRegistry()
	exporterRegistry.MustRegister(collectors.NewGoCollector())
	exporterRegistry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	// Register only exporter self-metrics, not device metrics
	exporterRegistry.MustRegister(exporter.Up)
	exporterRegistry.MustRegister(exporter.Duration)

	// Create registry for scraped device metrics (includes exporter metrics too)
	scrapeRegistry := prometheus.NewRegistry()
	scrapeRegistry.MustRegister(exporter)

	return exporterRegistry, scrapeRegistry
}

func setupServer(ctx context.Context, exporterRegistry, scrapeRegistry *prometheus.Registry) *http.Server {
	mux := http.NewServeMux()

	// Exporter metrics endpoint
	mux.Handle("/metrics", promhttp.HandlerFor(exporterRegistry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))

	// Scrape endpoint (returns scraped device metrics)
	mux.HandleFunc("/scrape", func(w http.ResponseWriter, r *http.Request) {
		// Gathering from the registry will automatically trigger Collect()
		// which performs the SNMP scrape
		handler := promhttp.HandlerFor(scrapeRegistry, promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		})
		handler.ServeHTTP(w, r)
	})

	// Root endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = io.WriteString(w, `<html>
<head><title>TP-Link DDM Exporter</title></head>
<body>
<h1>TP-Link DDM Exporter</h1>
<p><a href="/metrics">Exporter Metrics</a></p>
<p><a href="/scrape">Scrape Device Metrics</a></p>
</body>
</html>`)
	})

	return &http.Server{
		ReadHeaderTimeout: 1 * time.Second,
		ReadTimeout:       1 * time.Second,
		Handler:           mux,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}
}

func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	return slog.New(slog.NewTextHandler(os.Stderr, opts))
}
