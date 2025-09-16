package controller_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library/mocks"
	testutils "github.com/project/library/internal/usecase/library/test"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

// FIXME Необходимо перенести моки в сабтесты при использовании t.Parallel

func Test_GetAuthorInfo(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	logger, _ := zap.NewProduction()
	authorUseCase := mocks.NewMockAuthorUseCase(ctrl)
	bookUseCase := mocks.NewMockBooksUseCase(ctrl)
	service := controller.New(logger, bookUseCase, authorUseCase)
	ctx := t.Context()

	type args struct {
		ctx context.Context
		req *library.GetAuthorInfoRequest
	}

	tests := []struct {
		name        string
		args        args
		want        *library.GetAuthorInfoResponse
		wantErrCode codes.Code
		wantErr     error
		mocksUsed   bool
	}{
		{
			name: "get author info | valid request",
			args: args{
				ctx,
				&library.GetAuthorInfoRequest{
					Id: "7a948d89-108c-4133-be30-788bd453c0cd",
				},
			},
			want: &library.GetAuthorInfoResponse{
				Id:   "7a948d89-108c-4133-be30-788bd453c0cd",
				Name: "Author Name",
			},
			wantErr:     nil,
			wantErrCode: codes.OK,
			mocksUsed:   true,
		},

		{
			name: "get author info | invalid id",
			args: args{
				ctx,
				&library.GetAuthorInfoRequest{
					Id: "1aboba2",
				},
			},
			want:        nil,
			wantErrCode: codes.InvalidArgument,
			wantErr:     nil,
			mocksUsed:   false,
		},
		{
			name: "get author info | empty id",
			args: args{
				ctx,
				&library.GetAuthorInfoRequest{
					Id: "",
				},
			},
			want:        nil,
			wantErrCode: codes.InvalidArgument,
			wantErr:     nil,
			mocksUsed:   false,
		},
		{
			name: "get author info | author not found",
			args: args{
				ctx,
				&library.GetAuthorInfoRequest{
					Id: uuid.NewString(),
				},
			},
			want:        &library.GetAuthorInfoResponse{},
			wantErrCode: codes.NotFound,
			wantErr:     entity.ErrAuthorNotFound,
			mocksUsed:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.mocksUsed {
				authorUseCase.
					EXPECT().
					GetAuthorInfo(ctx, test.args.req.GetId()).
					Return(test.want.GetName(), test.wantErr)
			}

			got, err := service.GetAuthorInfo(test.args.ctx, test.args.req)

			testutils.CheckError(t, err, test.wantErrCode)
			if err == nil && test.want != nil {
				assert.Equal(t, test.want.Id, got.GetId())
				assert.Equal(t, test.want.Name, got.GetName())
			}
		})
	}
}
