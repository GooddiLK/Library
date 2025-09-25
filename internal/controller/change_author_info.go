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

func (i *impl) ChangeAuthorInfo(ctx context.Context, req *library.ChangeAuthorInfoRequest) (*library.ChangeAuthorInfoResponse, error) {
	tracer := otel.Tracer("library-service")
	ctx, span := tracer.Start(ctx, "ChangeAuthorInfo")
	defer span.End()

	entity.SendLoggerInfoWithCondition(i.logger, ctx, "Received ChangeAuthorInfo request.",
		layerCont, "author_id", req.GetId())

	if err := req.ValidateAll(); err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Code(codes.InvalidArgument), "Invalid ChangeAuthorInfo request.")
		i.logger.Error("Invalid ChangeAuthorInfo request.", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := i.authorUseCase.ChangeAuthor(ctx, req.GetId(), req.GetName())

	if err != nil {
		span.RecordError(err)
		i.logger.Error("Failed to change author info.", zap.Error(err))
		return nil, i.ConvertErr(err)
	}

	return &library.ChangeAuthorInfoResponse{}, nil
}
