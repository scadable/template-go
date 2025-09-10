package logger

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

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
