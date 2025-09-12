package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"template-go/internal/config"
)

// InitOtel initializes OpenTelemetry for tracing and metrics.
// It returns a shutdown function to gracefully close the providers.
func InitOtel(ctx context.Context, cfg config.Config) (func(context.Context) error, error) {
	// Create resource describing this service
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.OTELServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// --- TRACER SETUP ---
	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OTLP trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// --- METRICS SETUP ---
	metricExporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Prometheus metric exporter: %w", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(metricExporter),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	// --- SHUTDOWN ---
	shutdown := func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		var shutdownErr error
		if err := tp.Shutdown(ctx); err != nil {
			shutdownErr = fmt.Errorf("tracer shutdown error: %w", err)
		}
		if err := mp.Shutdown(ctx); err != nil {
			if shutdownErr != nil {
				shutdownErr = fmt.Errorf("%v; meter shutdown error: %w", shutdownErr, err)
			} else {
				shutdownErr = fmt.Errorf("meter shutdown error: %w", err)
			}
		}
		return shutdownErr
	}

	return shutdown, nil
}
