package controller

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

func Test_GetBookInfo(t *testing.T) {
	t.Parallel()
	logger, _ := zap.NewProduction()
	ctx := t.Context()

	type args struct {
		ctx context.Context
		req *library.GetBookInfoRequest
	}

	tests := []struct {
		name        string
		args        args
		want        *entity.Book
		wantErrCode codes.Code
		wantErr     error
		mocksUsed   bool
	}{
		{
			name: "get book info | valid request",
			args: args{
				ctx,
				&library.GetBookInfoRequest{
					Id: "7a948d89-108c-4133-be30-788bd453c0cd",
				},
			},
			want: &entity.Book{
				ID:        "7a948d89-108c-4133-be30-788bd453c0cd",
				Name:      "Test Book",
				AuthorIDs: []string{uuid.NewString(), uuid.NewString()},
			},
			wantErrCode: codes.OK,
			mocksUsed:   true,
		},
		{
			name: "get book info | invalid uuid",
			args: args{
				ctx,
				&library.GetBookInfoRequest{
					Id: "abobus12",
				},
			},
			wantErrCode: codes.InvalidArgument,
			mocksUsed:   false,
		},
		{
			name: "get book info | empty id",
			args: args{
				ctx,
				&library.GetBookInfoRequest{
					Id: "",
				},
			},
			wantErrCode: codes.InvalidArgument,
			mocksUsed:   false,
		},
		{
			name: "get book info | book not found",
			args: args{
				ctx,
				&library.GetBookInfoRequest{
					Id: "7a948d89-108c-4133-be30-788bd453c0cd",
				},
			},
			wantErrCode: codes.NotFound,
			wantErr:     entity.ErrBookNotFound,
			mocksUsed:   true,
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
				bookUseCase.
					EXPECT().
					GetBook(ctx, test.args.req.GetId()).
					Return(test.want, test.wantErr)
			}

			got, err := service.GetBookInfo(test.args.ctx, test.args.req)

			testutils.CheckError(t, err, test.wantErrCode)
			if err == nil {
				assert.Equal(t, test.want.ID, got.GetBook().GetId())
				assert.Equal(t, test.want.Name, got.GetBook().GetName())
				assert.Equal(t, test.want.AuthorIDs, got.GetBook().GetAuthorId())
			}
		})
	}
}
