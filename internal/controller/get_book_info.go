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

func (i *impl) GetBookInfo(ctx context.Context, req *library.GetBookInfoRequest) (*library.GetBookInfoResponse, error) {
	tracer := otel.Tracer("library-service")
	ctx, span := tracer.Start(ctx, "GetBookInfo")
	defer span.End()

	entity.SendLoggerInfoWithCondition(i.logger, ctx, "Received GetBookInfo request.",
		layerCont, "book_id", req.GetId())

	if err := req.ValidateAll(); err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Code(codes.InvalidArgument), "Invalid GetBookInfo request.")
		i.logger.Error("Invalid GetBookInfo request.", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	book, err := i.booksUseCase.GetBook(ctx, req.GetId())

	if err != nil {
		span.RecordError(err)
		i.logger.Error("Failed to get book.", zap.Error(err))
		return nil, i.ConvertErr(err)
	}

	return &library.GetBookInfoResponse{
		Book: &library.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.AuthorIDs,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		},
	}, nil
}
