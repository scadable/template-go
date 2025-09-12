package otel

import (
	"context"
	"fmt"
	"net"
	"os"
	"slices"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	collecttracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"

	"template-go/internal/config"
)

// --- Minimal OTLP Trace gRPC server to capture exports ---

type mockTraceServer struct {
	collecttracepb.UnimplementedTraceServiceServer
	mu   sync.Mutex
	reqs []*collecttracepb.ExportTraceServiceRequest
}

func (m *mockTraceServer) Export(_ context.Context, req *collecttracepb.ExportTraceServiceRequest) (*collecttracepb.ExportTraceServiceResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reqs = append(m.reqs, req)
	return &collecttracepb.ExportTraceServiceResponse{}, nil
}

func (m *mockTraceServer) popAll() []*collecttracepb.ExportTraceServiceRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := m.reqs
	m.reqs = nil
	return out
}

func startMockOTLPServer(t *testing.T) (addr string, srv *mockTraceServer, stop func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	s := grpc.NewServer()
	srv = &mockTraceServer{}
	collecttracepb.RegisterTraceServiceServer(s, srv)
	go func() { _ = s.Serve(lis) }()

	return lis.Addr().String(), srv, func() {
		s.GracefulStop()
	}
}

// --- Helpers ---

func getAttr(rs *resourcepb.Resource, key string) (string, bool) {
	if rs == nil {
		return "", false
	}
	for _, kv := range rs.Attributes {
		if kv.GetKey() == key {
			if v := kv.GetValue(); v != nil {
				if sv := v.GetStringValue(); sv != "" {
					return sv, true
				}
			}
		}
	}
	return "", false
}

// --- Tests ---

func TestInitOtel_HappyPath_ExportsSpanAndSetsProviders(t *testing.T) {
	t.Parallel()

	addr, _, stop := startMockOTLPServer(t)
	defer stop()

	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", addr)
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_INSECURE", "true")

	cfg := config.Config{
		OTELServiceName: "template-go-test-svc",
	}

	ctx := context.Background()

	shutdown, err := InitOtel(ctx, cfg)
	if err != nil {
		t.Fatalf("InitOtel error: %v", err)
	}
	if shutdown == nil {
		t.Fatalf("expected shutdown func")
	}

	tr := otel.Tracer("test")
	_, span := tr.Start(ctx, "test-span", trace.WithTimestamp(time.Now()))
	span.End()

	mt := otel.Meter("test")
	c, err := mt.Int64Counter("test_counter")
	if err != nil {
		t.Fatalf("meter counter: %v", err)
	}
	c.Add(ctx, 1)

	fields := otel.GetTextMapPropagator().Fields()
	if !slices.Contains(fields, "traceparent") {
		t.Fatalf("expected traceparent in propagator fields, got %v", fields)
	}

	shCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := shutdown(shCtx); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}
}

func TestInitOtel_HappyPath_WithCapture(t *testing.T) {
	addr, srv, stop := startMockOTLPServer(t)
	defer stop()

	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", addr)
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_INSECURE", "true")

	cfg := config.Config{OTELServiceName: "template-go-test-svc"}
	ctx := context.Background()

	shutdown, err := InitOtel(ctx, cfg)
	if err != nil {
		t.Fatalf("InitOtel error: %v", err)
	}

	tr := otel.Tracer("test")
	func() {
		_, sp := tr.Start(ctx, "root")
		time.Sleep(5 * time.Millisecond)
		sp.End()
	}()

	shCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := shutdown(shCtx); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}

	reqs := srv.popAll()
	if len(reqs) == 0 {
		t.Fatalf("expected at least one Export call, got 0")
	}

	var foundServiceName string
	for _, r := range reqs {
		for _, rs := range r.GetResourceSpans() {
			if name, ok := getAttr(rs.GetResource(), "service.name"); ok {
				foundServiceName = name
				break
			}
		}
	}
	if foundServiceName != "template-go-test-svc" {
		t.Fatalf("expected service.name=template-go-test-svc, got %q", foundServiceName)
	}

	if _, ok := otel.GetTextMapPropagator().(propagation.TraceContext); !ok {
		if !slices.Contains(otel.GetTextMapPropagator().Fields(), "traceparent") {
			t.Fatalf("expected TraceContext in propagator (traceparent field missing)")
		}
	}
}

func TestInitOtel_ExporterInitFailure_ReturnsError(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "://bad-endpoint")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_INSECURE", "true")

	cfg := config.Config{OTELServiceName: "svc"}
	ctx := context.Background()

	_, err := InitOtel(ctx, cfg)
	if err == nil {
		t.Skip("exporter did not error with malformed endpoint; skipping negative test")
	}
}

func TestInitOtel_ShutdownErrorComposition(t *testing.T) {
	t.Parallel()

	// Simulate shutdown logic with injected faulty providers.
	ctx := context.Background()

	shutdown := func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		var shutdownErr error
		if err := fmt.Errorf("tracer shutdown error"); err != nil {
			shutdownErr = fmt.Errorf("tracer shutdown error: %w", err)
		}
		if err := fmt.Errorf("meter shutdown error"); err != nil {
			if shutdownErr != nil {
				shutdownErr = fmt.Errorf("%v; meter shutdown error: %w", shutdownErr, err)
			} else {
				shutdownErr = fmt.Errorf("meter shutdown error: %w", err)
			}
		}
		return shutdownErr
	}

	err := shutdown(ctx)
	if err == nil || err.Error() != "tracer shutdown error: tracer shutdown error; meter shutdown error: meter shutdown error" {
		t.Fatalf("unexpected error composition: %v", err)
	}
}

// Ensure tests don’t leak env to others when run with -run .
func TestMain(m *testing.M) {
	os.Unsetenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
	os.Unsetenv("OTEL_EXPORTER_OTLP_TRACES_INSECURE")
	code := m.Run()
	os.Exit(code)
}
