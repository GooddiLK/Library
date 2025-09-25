package controller

import (
	"context"
	"github.com/project/library/internal/entity"
	"go.opentelemetry.io/otel"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/generated/api/library"
)

func (i *impl) RegisterAuthor(ctx context.Context, req *library.RegisterAuthorRequest) (*library.RegisterAuthorResponse, error) {
	tracer := otel.Tracer("library-service")
	ctx, span := tracer.Start(ctx, "RegisterAuthor")
	defer span.End()

	entity.SendLoggerInfoWithCondition(i.logger, ctx, "Received RegisterAuthor request.",
		layerCont, "author_name", req.GetName())

	if err := req.ValidateAll(); err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Code(codes.InvalidArgument), "Invalid RegisterAuthor request.")
		i.logger.Error("Invalid RegisterAuthor request.", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorUseCase.RegisterAuthor(ctx, req.GetName())

	if err != nil {
		span.RecordError(err)
		i.logger.Error("Failed to register author.", zap.Error(err))
		return nil, i.ConvertErr(err)
	}

	return &library.RegisterAuthorResponse{
		Id: author.ID,
	}, nil
}
