package controller

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/internal/entity"
)

func (i *impl) ConvertErr(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, entity.ErrAuthorNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, entity.ErrBookNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func SendAddBookLoggerInfo(logger *zap.Logger, ctx context.Context, message, arg1, arg2 string, strings []string) {
	logger.Info(message,
		zap.String("trace_id", trace.SpanFromContext(ctx).SpanContext().TraceID().String()),
		zap.String("layer", arg1),
		zap.String("book_name", arg2),
		zap.Strings("author_ids", strings),
	)
}

func CreateTracerSpan(ctx context.Context, spanMsg string) (context.Context, trace.Span) {
	tracer := otel.Tracer("library-service")
	return tracer.Start(ctx, spanMsg)
}

func SendSpanStatusLoggerError(logger *zap.Logger, ctx context.Context, message string, err error, code codes.Code) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)                         // Записывает факт ошибки
	span.SetStatus(otelCodes.Code(code), message) // Записывает чем завершилась операция
	logger.Error(message,
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.Error(err),
	)
}
