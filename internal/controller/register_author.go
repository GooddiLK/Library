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

	ctx, span := createTracerSpan(ctx, "RegisterAuthor")
	defer span.End()

	entity.SendLoggerInfoWithCondition(i.logger, ctx, "Received RegisterAuthor request.",
		layerCont, "author_name", req.GetName())

	if err := req.ValidateAll(); err != nil {
		SendSpanLoggerError(i.logger, ctx, "Invalid RegisterAuthor request.", err, codes.InvalidArgument)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorUseCase.RegisterAuthor(ctx, req.GetName())

	if err != nil {
		SendSpanLoggerError(i.logger, ctx, "Failed to register author.", err, codes.Internal)
		return nil, i.ConvertErr(err)
	}

	return &library.RegisterAuthorResponse{
		Id: author.ID,
	}, nil
}
