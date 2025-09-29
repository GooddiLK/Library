package controller

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
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

	tracer := otel.Tracer("library-service")
	ctx, span := tracer.Start(ctx, "AddBook")
	defer span.End()

	i.logger.Info("Received AddBook request.",
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.String("layer", layerCont),
		zap.String("book_name", req.GetName()),
		zap.Strings("author_ids", req.GetAuthorId()),
	)

	if err := req.ValidateAll(); err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Code(codes.InvalidArgument), "Invalid AddBook request.")
		i.logger.Error("Invalid AddBook request.", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	book, err := i.booksUseCase.AddBook(ctx, req.GetName(), req.GetAuthorId())

	if err != nil {
		span.RecordError(err)
		i.logger.Error("Failed to add book.", zap.Error(err))
		return nil, i.ConvertErr(err)
	}

	return &library.AddBookResponse{
		Book: &library.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.AuthorIDs,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		},
	}, nil
}
