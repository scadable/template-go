package logger

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// A thread-safe buffer is used to capture log output for assertions.
type syncer struct {
	sync.Mutex
	bytes.Buffer
}

// Sync implements the zapcore.WriteSyncer interface.
func (s *syncer) Sync() error {
	return nil
}

// setupTestLogger configures a temporary global logger that writes to an in-memory buffer.
func setupTestLogger(buffer *syncer) {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		buffer,
		zap.DebugLevel,
	)
	zapLogger := zap.New(core)
	log = zapLogger
}

func TestInit(t *testing.T) {
	// This test ensures that the logger initialization does not panic.
	assert.NotPanics(t, func() {
		Init()
	})
}

func TestInitPanicsIfLoggerFails(t *testing.T) {
	// Backup the real logger factory
	original := newLogger
	defer func() { newLogger = original }()

	// Simulate failure
	newLogger = func(_ ...zap.Option) (*zap.Logger, error) {
		return nil, assert.AnError
	}
	
	assert.PanicsWithValue(t,
		"cannot initialize zap logger: assert.AnError general error for testing",
		func() {
			Init()
		},
	)
}

func TestLogLevelsAndSync(t *testing.T) {
	// This test ensures all log level functions write messages correctly and Sync can be called.
	var buffer syncer
	setupTestLogger(&buffer)

	ctx := context.Background()

	// Act: Log at different levels
	Debug(ctx, "debug message")
	Info(ctx, "info message")
	Warn(ctx, "warn message")
	Error(ctx, "error message")
	Sync() // Also covers the Sync() function

	output := buffer.String()

	// Assert: Check that all messages were logged
	assert.Contains(t, output, "debug message")
	assert.Contains(t, output, "info message")
	assert.Contains(t, output, "warn message")
	assert.Contains(t, output, "error message")
}

func TestInjectTrace(t *testing.T) {
	// This test ensures trace and span IDs are correctly injected into logs.
	t.Run("should do nothing if context has no trace", func(t *testing.T) {
		// This sub-test covers the case where the context is empty.
		ctx := context.Background()
		fields := []zap.Field{}

		result := injectTrace(ctx, fields)

		assert.Empty(t, result, "Expected no fields to be added")
	})

	t.Run("should add trace and span IDs if context has trace", func(t *testing.T) {
		// This sub-test covers the case where the context contains trace info.
		traceID, _ := trace.TraceIDFromHex("8c3c1e95c9a0989f4e42a448557b4914")
		spanID, _ := trace.SpanIDFromHex("8177353f49635e07")

		spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: traceID,
			SpanID:  spanID,
		})
		ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)
		fields := []zap.Field{}

		result := injectTrace(ctx, fields)

		require.Len(t, result, 2)
		assert.Equal(t, "trace_id", result[0].Key)
		assert.Equal(t, "8c3c1e95c9a0989f4e42a448557b4914", result[0].String)
		assert.Equal(t, "span_id", result[1].Key)
		assert.Equal(t, "8177353f49635e07", result[1].String)
	})
}
