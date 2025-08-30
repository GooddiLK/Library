package controller_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/usecase/library/mocks"
	testutils "github.com/project/library/internal/usecase/library/test"
)

// FIXME Необходимо перенести моки в сабтесты при использовании t.Parallel

func Test_ChangeAuthorInfo(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	logger, _ := zap.NewProduction()
	authorUseCase := mocks.NewMockAuthorUseCase(ctrl)
	bookUseCase := mocks.NewMockBooksUseCase(ctrl)
	service := controller.New(logger, bookUseCase, authorUseCase)
	ctx := t.Context()

	type args struct {
		ctx context.Context
		req *library.ChangeAuthorInfoRequest
	}

	tests := []struct {
		name        string
		args        args
		wantErrCode codes.Code
		wantErr     error
		mocksUsed   bool
	}{
		{
			"change author info | without error",
			args{ctx,
				&library.ChangeAuthorInfoRequest{
					Id:   uuid.NewString(),
					Name: "New Name",
				},
			},
			codes.OK,
			nil,
			true,
		},

		{
			"change author info | with uncorrected Id",
			args{ctx,
				&library.ChangeAuthorInfoRequest{
					Id:   "10",
					Name: "New Name",
				},
			},
			codes.InvalidArgument,
			status.Error(codes.InvalidArgument, " uncorrected id"),
			false,
		},

		{
			"change author info | with invalid name",
			args{
				ctx,
				&library.ChangeAuthorInfoRequest{
					Id:   uuid.NewString(),
					Name: "",
				},
			},
			codes.InvalidArgument,
			status.Error(codes.InvalidArgument, " invalid author name "),
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.mocksUsed {
				authorUseCase.EXPECT().
					UpdateAuthor(ctx, test.args.req.GetId(), test.args.req.GetName()).
					Return(test.wantErr)
			}

			_, err := service.ChangeAuthorInfo(test.args.ctx, test.args.req)

			testutils.CheckError(t, err, test.wantErrCode)
		})
	}
}
