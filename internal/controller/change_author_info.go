package controller

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/generated/api/library"
)

func (i *impl) ChangeAuthorInfo(ctx context.Context, req *library.ChangeAuthorInfoRequest) (*library.ChangeAuthorInfoResponse, error) {
	i.logger.Info("Received ChangeAuthorInfo request.",
		zap.String("new_author_name", req.GetName()),
		zap.String("author_id", req.GetId()))

	if err := req.ValidateAll(); err != nil {
		i.logger.Error("Invalid ChangeAuthorInfo request.", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := i.authorUseCase.ChangeAuthor(ctx, req.GetId(), req.GetName())

	if err != nil {
		i.logger.Error("Failed to change author info.", zap.Error(err))
		return nil, i.ConvertErr(err)
	}

	return &library.ChangeAuthorInfoResponse{}, nil
}
