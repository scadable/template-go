package otel

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"template-go/internal/config"
)

// --- Test Cases for Initialization Errors ---

func TestInitOtel_ResourceError(t *testing.T) {
	// GIVEN a mock resource constructor that fails
	original := newResource
	defer func() { newResource = original }()
	newResource = func(ctx context.Context, opts ...resource.Option) (*resource.Resource, error) {
		return nil, errors.New("resource creation failed")
	}

	// WHEN InitOtel is called
	_, err := InitOtel(context.Background(), config.Config{OTELServiceName: "test"})

	// THEN the correct error is returned
	if err == nil || !strings.Contains(err.Error(), "failed to create resource") {
		t.Fatalf("expected resource creation error, got: %v", err)
	}
}

func TestInitOtel_TraceExporterError(t *testing.T) {
	// GIVEN a mock trace exporter constructor that fails
	old := newTraceExporter
	defer func() { newTraceExporter = old }()
	newTraceExporter = func(ctx context.Context, opts ...otlptracegrpc.Option) (*otlptrace.Exporter, error) {
		return nil, errors.New("trace exporter fail")
	}

	// WHEN InitOtel is called
	_, err := InitOtel(context.Background(), config.Config{OTELServiceName: "test"})

	// THEN the correct error is returned
	if err == nil || !strings.Contains(err.Error(), "failed to initialize OTLP trace exporter") {
		t.Fatalf("expected trace exporter error, got: %v", err)
	}
}

func TestInitOtel_MetricExporterError(t *testing.T) {
	// GIVEN a mock metric exporter constructor that fails
	old := newPromExporter
	defer func() { newPromExporter = old }()
	newPromExporter = func(opts ...prometheus.Option) (*prometheus.Exporter, error) {
		return nil, errors.New("metric exporter fail")
	}

	// WHEN InitOtel is called
	_, err := InitOtel(context.Background(), config.Config{OTELServiceName: "test"})

	// THEN the correct error is returned
	if err == nil || !strings.Contains(err.Error(), "failed to initialize Prometheus metric exporter") {
		t.Fatalf("expected metric exporter error, got: %v", err)
	}
}

// --- Test Case for Successful Initialization and Shutdown ---

func TestInitOtel_SuccessfulInitAndShutdown(t *testing.T) {
	// GIVEN mocks for successful initialization
	oldTracerProv := newTracerProvider
	oldMeterProv := newMeterProvider
	defer func() {
		newTracerProvider = oldTracerProv
		newMeterProvider = oldMeterProv
	}()
	// Use the real constructors to get properly initialized, non-nil providers
	newTracerProvider = func(opts ...sdktrace.TracerProviderOption) tracerProvider {
		return sdktrace.NewTracerProvider(opts...)
	}
	newMeterProvider = func(opts ...metric.Option) meterProvider {
		return metric.NewMeterProvider(opts...)
	}

	// WHEN InitOtel is called and shutdown is executed
	shutdown, err := InitOtel(context.Background(), config.Config{OTELServiceName: "test"})

	// THEN initialization and shutdown succeed without errors
	if err != nil {
		t.Fatalf("unexpected error during InitOtel: %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected shutdown function, got nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		t.Fatalf("expected no shutdown error, got: %v", err)
	}
}

// --- Mocks and Test Cases for Shutdown Errors ---

// mockFailingTracerProvider implements our local `tracerProvider` interface to test shutdown errors.
type mockFailingTracerProvider struct{}

func (m *mockFailingTracerProvider) Shutdown(ctx context.Context) error {
	return errors.New("tracer shutdown failed")
}

// mockFailingMeterProvider implements our local `meterProvider` interface to test shutdown errors.
type mockFailingMeterProvider struct{}

func (m *mockFailingMeterProvider) Shutdown(ctx context.Context) error {
	return errors.New("meter shutdown failed")
}

func TestInitOtel_ShutdownTracerErrorOnly(t *testing.T) {
	// GIVEN a tracer provider that fails on shutdown
	oldTracerProv := newTracerProvider
	defer func() { newTracerProvider = oldTracerProv }()
	newTracerProvider = func(opts ...sdktrace.TracerProviderOption) tracerProvider {
		return &mockFailingTracerProvider{}
	}

	// WHEN the shutdown function is called
	shutdown, err := InitOtel(context.Background(), config.Config{OTELServiceName: "test"})
	if err != nil {
		t.Fatalf("unexpected error during init: %v", err)
	}
	err = shutdown(context.Background())

	// THEN the correct error is returned
	if err == nil || !strings.Contains(err.Error(), "tracer shutdown error") {
		t.Fatalf("expected tracer shutdown error, got: %v", err)
	}
}

func TestInitOtel_ShutdownMeterErrorOnly(t *testing.T) {
	// GIVEN a meter provider that fails on shutdown
	oldMeterProv := newMeterProvider
	defer func() { newMeterProvider = oldMeterProv }()
	newMeterProvider = func(opts ...metric.Option) meterProvider {
		return &mockFailingMeterProvider{}
	}

	// WHEN the shutdown function is called
	shutdown, err := InitOtel(context.Background(), config.Config{OTELServiceName: "test"})
	if err != nil {
		t.Fatalf("unexpected error during init: %v", err)
	}
	err = shutdown(context.Background())

	// THEN the correct error is returned, specifically testing the `else` branch of error aggregation
	if err == nil || !strings.Contains(err.Error(), "meter shutdown error") {
		t.Fatalf("expected meter shutdown error, got: %v", err)
	}
}

func TestInitOtel_ShutdownBothError(t *testing.T) {
	// GIVEN both providers fail on shutdown
	oldTracerProv := newTracerProvider
	oldMeterProv := newMeterProvider
	defer func() {
		newTracerProvider = oldTracerProv
		newMeterProvider = oldMeterProv
	}()
	newTracerProvider = func(opts ...sdktrace.TracerProviderOption) tracerProvider {
		return &mockFailingTracerProvider{}
	}
	newMeterProvider = func(opts ...metric.Option) meterProvider {
		return &mockFailingMeterProvider{}
	}

	// WHEN the shutdown function is called
	shutdown, err := InitOtel(context.Background(), config.Config{OTELServiceName: "test"})
	if err != nil {
		t.Fatalf("unexpected error during init: %v", err)
	}
	err = shutdown(context.Background())

	// THEN both errors are aggregated in the final error message
	if err == nil || !strings.Contains(err.Error(), "tracer shutdown error") || !strings.Contains(err.Error(), "meter shutdown error") {
		t.Fatalf("expected both shutdown errors, got: %v", err)
	}
}
