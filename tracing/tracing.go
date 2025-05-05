// Package tracing provides OpenTelemetry distributed tracing functionality
package tracing

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerProvider is a global variable that provides tracers
var TracerProvider *sdktrace.TracerProvider

// Initialize sets up OpenTelemetry with the OTLP exporter
func Initialize(ctx context.Context) (shutdown func(context.Context) error, err error) {
	// Determine service name from environment or use default
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "parsel-unknown-service"
	}

	// Get the collector endpoint from environment or use default
	collectorEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if collectorEndpoint == "" {
		// Try to connect to Tempo in the monitoring network via host.docker.internal
		collectorEndpoint = "host.docker.internal:4317" // Use host.docker.internal to reach Tempo
	}

	// Create the OTLP exporter
	exporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(collectorEndpoint),
			otlptracegrpc.WithInsecure(), // Remove this in production and use proper TLS
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	// Create a resource describing this application
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
			attribute.String("environment", os.Getenv("APP_ENV")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating resource: %w", err)
	}

	// Configure the trace provider
	TracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // In production, use an appropriate sampler
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set the global trace provider
	otel.SetTracerProvider(TracerProvider)

	// Set the global propagator to tracecontext (the default)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // W3C Trace Context format
		propagation.Baggage{},      // W3C Baggage format
	))

	// Return a function that can be called to clean up resources when shutting down
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := TracerProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutting down tracer provider: %w", err)
		}
		return nil
	}, nil
}

// Tracer returns a named tracer from the global provider
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// SpanFromContext returns the current span from a context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// ContextWithSpan adds a span to a context
func ContextWithSpan(ctx context.Context, span trace.Span) context.Context {
	return trace.ContextWithSpan(ctx, span)
}
