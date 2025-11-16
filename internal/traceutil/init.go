package traceutil

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Init initializes OpenTelemetry tracing if OTEL environment variables are set.
// Returns a shutdown function if tracing was initialized, nil otherwise.
func Init(ctx context.Context) (func(context.Context) error, error) {
	// Check if tracing is disabled
	if os.Getenv("OTEL_SDK_DISABLED") == "true" {
		return nil, nil
	}

	// Check if OTLP endpoint is configured
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
	if endpoint == "" {
		endpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	}

	if endpoint == "" {
		// No tracing configured
		return nil, nil
	}

	// Create OTLP exporter
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
	}

	// Check if TLS is disabled
	if os.Getenv("OTEL_EXPORTER_OTLP_TRACES_INSECURE") == "true" || os.Getenv("OTEL_EXPORTER_OTLP_INSECURE") == "true" {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("tplink-ddm-exporter"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Return shutdown function
	return tp.Shutdown, nil
}

// Tracer returns the global tracer.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
