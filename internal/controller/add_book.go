package controller

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"

	"github.com/project/library/generated/api/library"
)

var (
	AddBookDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "library_add_book_duration_ms",
		Help:    "Duration of AddBook in ms", // Комментарий для http://localhost:9000/metrics
		Buckets: prometheus.DefBuckets,
	})

	AddBookRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "library_add_book_requests_total",
		Help: "Total number of AddBook requests",
	})
)

func init() {
	prometheus.MustRegister(AddBookDuration)
	prometheus.MustRegister(AddBookRequests)
}

func (i *impl) AddBook(ctx context.Context, req *library.AddBookRequest) (*library.AddBookResponse, error) {
	AddBookRequests.Inc()
	start := time.Now()
	defer func() {
		AddBookDuration.Observe(float64(time.Since(start).Milliseconds()))
	}()

	ctx, span := CreateTracerSpan(ctx, "AddBook")
	defer span.End()

	SendAddBookLoggerInfo(i.logger, ctx, "Received AddBook request.",
		layerCont, req.GetName(), req.GetAuthorIds())

	if err := req.ValidateAll(); err != nil {
		SendSpanStatusLoggerError(i.logger, ctx, "Invalid AddBook request.", err, codes.InvalidArgument)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	book, err := i.booksUseCase.AddBook(ctx, req.GetName(), req.GetAuthorIds())

	if err != nil {
		SendSpanStatusLoggerError(i.logger, ctx, "Failed to add book.", err, codes.Internal)
		return nil, i.ConvertErr(err)
	}

	return &library.AddBookResponse{
		Book: &library.Book{
			Id:        book.Id,
			Name:      book.Name,
			AuthorIds: book.AuthorIds,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		},
	}, nil
}
