package controller

import (
	"context"
	"github.com/project/library/internal/entity"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"

	"github.com/project/library/generated/api/library"
)

var (
	RegisterAuthorDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "library_register_author_duration_ms",
		Help:    "Duration of RegisterAuthor in ms",
		Buckets: prometheus.DefBuckets,
	})

	RegisterAuthorRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "library_register_author_requests_total",
		Help: "Total number of RegisterAuthor requests",
	})
)

func init() {
	prometheus.MustRegister(RegisterAuthorDuration)
	prometheus.MustRegister(RegisterAuthorRequests)
}

func (i *impl) RegisterAuthor(ctx context.Context, req *library.RegisterAuthorRequest) (*library.RegisterAuthorResponse, error) {
	RegisterAuthorRequests.Inc()
	start := time.Now()
	defer func() {
		RegisterAuthorDuration.Observe(float64(time.Since(start).Milliseconds()))
	}()

	tracer := otel.Tracer("library-service")
	ctx, span := tracer.Start(ctx, "RegisterAuthor")
	defer span.End()

	entity.SendLoggerInfoWithCondition(i.logger, ctx, "Received RegisterAuthor request.",
		layerCont, "author_name", req.GetName())

	if err := req.ValidateAll(); err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Code(codes.InvalidArgument), "Invalid RegisterAuthor request.")
		i.logger.Error("Invalid RegisterAuthor request.", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorUseCase.RegisterAuthor(ctx, req.GetName())

	if err != nil {
		span.RecordError(err)
		i.logger.Error("Failed to register author.", zap.Error(err))
		return nil, i.ConvertErr(err)
	}

	return &library.RegisterAuthorResponse{
		Id: author.ID,
	}, nil
}
