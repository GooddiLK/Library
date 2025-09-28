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
	GetAuthorInfoDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "library_get_author_info_duration_ms",
		Help:    "Duration of GetAuthorInfo in ms",
		Buckets: prometheus.DefBuckets,
	})

	GetAuthorInfoRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "library_get_author_info_requests_total",
		Help: "Total number of GetAuthorInfo requests",
	})
)

func init() {
	prometheus.MustRegister(GetAuthorInfoDuration)
	prometheus.MustRegister(GetAuthorInfoRequests)
}

func (i *impl) GetAuthorInfo(ctx context.Context, req *library.GetAuthorInfoRequest) (*library.GetAuthorInfoResponse, error) {
	GetAuthorInfoRequests.Inc()
	start := time.Now()
	defer func() {
		GetAuthorInfoDuration.Observe(float64(time.Since(start).Milliseconds()))
	}()

	tracer := otel.Tracer("library-service")
	ctx, span := tracer.Start(ctx, "GetAuthorInfo")
	defer span.End()

	entity.SendLoggerInfoWithCondition(i.logger, ctx, "Received GetAuthorInfo request.",
		layerCont, "author_id", req.GetId())

	if err := req.ValidateAll(); err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Code(codes.InvalidArgument), "Invalid GetAuthorInfo request.")
		i.logger.Error("Invalid GetAuthorInfo request.", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorUseCase.GetAuthorInfo(ctx, req.GetId())

	if err != nil {
		span.RecordError(err)
		i.logger.Error("Failed to get author info.", zap.Error(err))
		return nil, i.ConvertErr(err)
	}

	return &library.GetAuthorInfoResponse{
		Id:   author.ID,
		Name: author.Name,
	}, nil
}
