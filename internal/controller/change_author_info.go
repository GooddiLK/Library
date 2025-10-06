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
	ChangeAuthorInfoDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "library_change_author_info_duration_ms",
		Help:    "Duration of ChangeAuthorInfo in ms",
		Buckets: prometheus.DefBuckets,
	})

	ChangeAuthorInfoRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "library_change_author_info_requests_total",
		Help: "Total number of ChangeAuthorInfo requests",
	})
)

func init() {
	prometheus.MustRegister(ChangeAuthorInfoDuration)
	prometheus.MustRegister(ChangeAuthorInfoRequests)
}

func (i *impl) ChangeAuthorInfo(ctx context.Context, req *library.ChangeAuthorInfoRequest) (*library.ChangeAuthorInfoResponse, error) {
	ChangeAuthorInfoRequests.Inc()
	start := time.Now()
	defer func() {
		ChangeAuthorInfoDuration.Observe(float64(time.Since(start).Milliseconds()))
	}()

	ctx, span := createTracerSpan(ctx, "ChangeAuthorInfo")
	defer span.End()

	entity.SendLoggerInfoWithCondition(i.logger, ctx, "Received ChangeAuthorInfo request.",
		layerCont, "author_id", req.GetId())

	if err := req.ValidateAll(); err != nil {
		SendSpanLoggerError(i.logger, ctx, "Invalid ChangeAuthorInfo request.", err, codes.InvalidArgument)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := i.authorUseCase.ChangeAuthor(ctx, req.GetId(), req.GetName())

	if err != nil {
		SendSpanLoggerError(i.logger, ctx, "Failed to change author info.", err, codes.Internal)
		return nil, i.ConvertErr(err)
	}

	return &library.ChangeAuthorInfoResponse{}, nil
}
