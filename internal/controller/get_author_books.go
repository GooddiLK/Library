package controller

import (
	"github.com/project/library/internal/entity"
	"go.opentelemetry.io/otel"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/project/library/generated/api/library"
)

func (i *impl) GetAuthorBooks(req *library.GetAuthorBooksRequest, server library.Library_GetAuthorBooksServer) error {
	ctx := server.Context()

	tracer := otel.Tracer("library-service")
	ctx, span := tracer.Start(ctx, "GetAuthorBooks")
	defer span.End()

	entity.SendLoggerInfoWithCondition(i.logger, ctx, "Received GetAuthorBooks request.",
		layerCont, "author_id", req.GetAuthorId())

	if err := req.ValidateAll(); err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Code(codes.InvalidArgument), "Invalid GetAuthorBooks request.")
		i.logger.Error("Invalid GetAuthorBooks request.", zap.Error(err))
		return status.Error(codes.InvalidArgument, err.Error())
	}

	books, err := i.booksUseCase.GetAuthorBooks(ctx, req.GetAuthorId())

	if err != nil {
		span.RecordError(err)
		i.logger.Error("Failed to get author books.", zap.Error(err))
		return i.ConvertErr(err)
	}

	for _, book := range books {
		err := server.Send(&library.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.AuthorIDs,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
