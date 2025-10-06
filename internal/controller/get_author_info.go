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

	ctx, span := CreateTracerSpan(ctx, "GetAuthorInfo")
	defer span.End()

	entity.SendLoggerInfoWithCondition(i.logger, ctx, "Received GetAuthorInfo request.",
		layerCont, "author_id", req.GetId())

	if err := req.ValidateAll(); err != nil {
		SendSpanStatusLoggerError(i.logger, ctx, "Invalid GetAuthorInfo request.", err, codes.InvalidArgument)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorUseCase.GetAuthorInfo(ctx, req.GetId())

	if err != nil {
		SendSpanStatusLoggerError(i.logger, ctx, "Failed to get author info.", err, codes.Internal)
		return nil, i.ConvertErr(err)
	}

	return &library.GetAuthorInfoResponse{
		Id:   author.Id,
		Name: author.Name,
	}, nil
}
