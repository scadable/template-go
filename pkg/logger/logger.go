package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var log *zap.Logger

// Init initializes the global production logger.
func Init() {
	var err error
	// zap.NewProduction() provides a sane, performant logger for production.
	log, err = zap.NewProduction()
	if err != nil {
		panic("cannot initialize zap logger: " + err.Error())
	}
}

// Sync flushes any buffered log entries.
func Sync() {
	// It's a good practice to call this before the application exits.
	_ = log.Sync()
}

// getLogger returns the global logger instance.
func getLogger() *zap.Logger {
	return log
}

// Info logs a message at the info level.
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	getLogger().Info(msg, injectTrace(ctx, fields)...)
}

// Error logs a message at the error level.
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	getLogger().Error(msg, injectTrace(ctx, fields)...)
}

// Debug logs a message at the debug level.
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	getLogger().Debug(msg, injectTrace(ctx, fields)...)
}

// Warn logs a message at the warn level.
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	getLogger().Warn(msg, injectTrace(ctx, fields)...)
}

// injectTrace checks for a trace in the context and adds trace_id and span_id
// to the log fields if found.
func injectTrace(ctx context.Context, fields []zap.Field) []zap.Field {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		fields = append(fields,
			zap.String("trace_id", spanCtx.TraceID().String()),
			zap.String("span_id", spanCtx.SpanID().String()),
		)
	}
	return fields
}
