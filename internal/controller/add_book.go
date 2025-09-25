package controller

import (
	"context"
	"github.com/project/library/internal/entity"
	"go.opentelemetry.io/otel"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/project/library/generated/api/library"
)

func (i *impl) AddBook(ctx context.Context, req *library.AddBookRequest) (*library.AddBookResponse, error) {
	tracer := otel.Tracer("library-service")
	ctx, span := tracer.Start(ctx, "AddBook")
	defer span.End()

	entity.SendLoggerInfoWithCondition(i.logger, ctx, "Received AddBook request.",
		layerCont, "book_name", req.GetName())

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
