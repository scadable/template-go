package logger

import (
	"context"
	"go.uber.org/zap"
)

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	getLogger().Info(msg, injectTrace(ctx, fields)...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	getLogger().Error(msg, injectTrace(ctx, fields)...)
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	getLogger().Debug(msg, injectTrace(ctx, fields)...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	getLogger().Warn(msg, injectTrace(ctx, fields)...)
}
