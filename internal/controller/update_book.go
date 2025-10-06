package controller

import (
	"context"
	"time"

	"github.com/project/library/internal/entity"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

	ctx, span := createTracerSpan(ctx, "UpdateBook")
	defer span.End()

	entity.SendLoggerInfoWithCondition(i.logger, ctx, "Received UpdateBook request.",
		layerCont, "book_id", req.GetId())

	if err := req.ValidateAll(); err != nil {
		SendSpanStatusLoggerError(i.logger, ctx, "Invalid UpdateBook request.", err, codes.InvalidArgument)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := i.booksUseCase.UpdateBook(ctx, req.GetId(), req.GetName(), req.GetAuthorIds())

	if err != nil {
		SendSpanStatusLoggerError(i.logger, ctx, "Failed to update book.", err, codes.Internal)
		return nil, i.ConvertErr(err)
	}

	return &library.UpdateBookResponse{}, nil
}
