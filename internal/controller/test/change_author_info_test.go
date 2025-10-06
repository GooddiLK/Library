package controller

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/usecase/library/mocks"
	testutils "github.com/project/library/internal/usecase/library/test"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_ChangeAuthorInfo(t *testing.T) {
	t.Parallel()
	logger, _ := zap.NewProduction()
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
		test := test // capture range variable
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Создаем моки внутри каждого субтеста
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authorUseCase := mocks.NewMockAuthorUseCase(ctrl)
			bookUseCase := mocks.NewMockBooksUseCase(ctrl)
			service := controller.New(logger, bookUseCase, authorUseCase)

			if test.mocksUsed {
				authorUseCase.EXPECT().
					ChangeAuthor(ctx, test.args.req.GetId(), test.args.req.GetName()).
					Return(test.wantErr)
			}

			_, err := service.ChangeAuthorInfo(test.args.ctx, test.args.req)

			testutils.CheckError(t, err, test.wantErrCode)
		})
	}
}
