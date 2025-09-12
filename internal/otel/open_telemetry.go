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

// --- Local interfaces for improved testability ---
type tracerProvider interface {
	Shutdown(context.Context) error
}

type meterProvider interface {
	Shutdown(context.Context) error
}

// --- Package-level constructors for testability ---
var (
	newResource       = resource.New
	newTraceExporter  = otlptracegrpc.New
	newPromExporter   = prometheus.New
	newTracerProvider = func(opts ...sdktrace.TracerProviderOption) tracerProvider {
		return sdktrace.NewTracerProvider(opts...)
	}
	newMeterProvider = func(opts ...metric.Option) meterProvider {
		return metric.NewMeterProvider(opts...)
	}
)

// InitOtel initializes OpenTelemetry for tracing and metrics.
func InitOtel(ctx context.Context, cfg config.Config) (func(context.Context) error, error) {
	res, err := newResource(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.OTELServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	traceExporter, err := newTraceExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OTLP trace exporter: %w", err)
	}

	// The type of `tp` is now our local `tracerProvider` interface
	tp := newTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	// We need to cast back to the concrete type for the global setter
	if realTP, ok := tp.(*sdktrace.TracerProvider); ok {
		otel.SetTracerProvider(realTP)
	}
	otel.SetTextMapPropagator(propagation.TraceContext{})

	metricExporter, err := newPromExporter()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Prometheus metric exporter: %w", err)
	}

	// The type of `mp` is now our local `meterProvider` interface
	mp := newMeterProvider(
		metric.WithReader(metricExporter),
		metric.WithResource(res),
	)
	// We need to cast back to the concrete type for the global setter
	if realMP, ok := mp.(*metric.MeterProvider); ok {
		otel.SetMeterProvider(realMP)
	}

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
