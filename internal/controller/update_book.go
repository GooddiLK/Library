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
	UpdateBookDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "library_update_book_duration_ms",
		Help:    "Duration of UpdateBook in ms",
		Buckets: prometheus.DefBuckets,
	})

	UpdateBookRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "library_update_book_requests_total",
		Help: "Total number of UpdateBook requests",
	})
)

func init() {
	prometheus.MustRegister(UpdateBookDuration)
	prometheus.MustRegister(UpdateBookRequests)
}

func (i *impl) UpdateBook(ctx context.Context, req *library.UpdateBookRequest) (*library.UpdateBookResponse, error) {
	UpdateBookRequests.Inc()
	start := time.Now()
	defer func() {
		UpdateBookDuration.Observe(float64(time.Since(start).Milliseconds()))
	}()

	tracer := otel.Tracer("library-service")
	ctx, span := tracer.Start(ctx, "UpdateBook")
	defer span.End()

	entity.SendLoggerInfoWithCondition(i.logger, ctx, "Received UpdateBook request.",
		layerCont, "book_id", req.GetId())

	if err := req.ValidateAll(); err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Code(codes.InvalidArgument), "Invalid UpdateBook request.")
		i.logger.Error("Invalid UpdateBook request.", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := i.booksUseCase.UpdateBook(ctx, req.GetId(), req.GetName(), req.GetAuthorIds())

	if err != nil {
		span.RecordError(err)
		i.logger.Error("Failed to update book.", zap.Error(err))
		return nil, i.ConvertErr(err)
	}

	return &library.UpdateBookResponse{}, nil
}
