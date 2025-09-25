package app

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
)

func initTracer(logger *zap.Logger, url string) func(context.Context) error {
	logger.Info("Starting tracer server.", zap.String("address", url))

	// Jaeger уже поднят по URL: - отсылай туда trace
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))

	if err != nil {
		logger.Fatal("Can not create jaeger collector.", zap.Error(err))
	}

	// Инициализация trace provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("library-service"),
		)),
	)

	otel.SetTracerProvider(tp)

	return tp.Shutdown
}
