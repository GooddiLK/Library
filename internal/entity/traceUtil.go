package entity

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func SendLoggerInfo(logger *zap.Logger, ctx context.Context, message string, layer string) {
	logger.Info(message,
		zap.String("trace_id", trace.SpanFromContext(ctx).SpanContext().TraceID().String()),
		zap.String("layer", layer))
}

func SendLoggerInfoWithCondition(logger *zap.Logger, ctx context.Context, message, layer, key, condition string) {
	logger.Info(message,
		zap.String("trace_id", trace.SpanFromContext(ctx).SpanContext().TraceID().String()),
		zap.String("layer", layer),
		zap.String(key, condition))
}

func SendLoggerSpanError(logger *zap.Logger, ctx context.Context, message string, layer string, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(fmt.Errorf(message+" %w", err))
	logger.Error(message,
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.Error(err))
}
