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

func Test_UpdateBook(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	logger, _ := zap.NewProduction()
	authorUseCase := mocks.NewMockAuthorUseCase(ctrl)
	bookUseCase := mocks.NewMockBooksUseCase(ctrl)
	service := controller.New(logger, bookUseCase, authorUseCase)
	ctx := t.Context()

	type args struct {
		ctx context.Context
		req *library.UpdateBookRequest
	}

	tests := []struct {
		name        string
		args        args
		wantErrCode codes.Code
		wantErr     error
		mocksUsed   bool
	}{
		{
			name: "update book | valid request with name and authors",
			args: args{
				ctx,
				&library.UpdateBookRequest{
					Id:       "7a948d89-108c-4133-be30-788bd453c0cd",
					Name:     "New name",
					AuthorId: []string{uuid.NewString(), uuid.NewString()},
				},
			},
			wantErrCode: codes.OK,
			wantErr:     nil,
			mocksUsed:   true,
		},
		{
			name: "update book | valid request with name only",
			args: args{
				ctx,
				&library.UpdateBookRequest{
					Id:   uuid.NewString(),
					Name: "New name",
				},
			},
			wantErrCode: codes.OK,
			mocksUsed:   true,
		},
		{
			name: "update book | invalid request with authors only",
			args: args{
				ctx,
				&library.UpdateBookRequest{
					Id:       "7a948d89-108c-4133-be30-788bd453c0cd",
					AuthorId: []string{"author-id-1"},
				},
			},
			wantErrCode: codes.InvalidArgument,
			mocksUsed:   false,
		},
		{
			name: "update book | invalid uuid",
			args: args{
				ctx,
				&library.UpdateBookRequest{
					Id: "aboba",
				},
			},
			wantErrCode: codes.InvalidArgument,
			mocksUsed:   false,
		},
		{
			name: "update book | empty id",
			args: args{
				ctx,
				&library.UpdateBookRequest{
					Id: "",
				},
			},
			wantErrCode: codes.InvalidArgument,
			mocksUsed:   false,
		},
		{
			name: "update book | invalid author ids",
			args: args{
				ctx,
				&library.UpdateBookRequest{
					Id:       "7a948d89-108c-4133-be30-788bd453c0cd",
					AuthorId: []string{"aboba"},
				},
			},
			wantErrCode: codes.InvalidArgument,
			mocksUsed:   false,
		},
		{
			name: "update book | book not found",
			args: args{
				ctx,
				&library.UpdateBookRequest{
					Id:   "7a948d89-108c-4133-be30-788bd453c0cd",
					Name: "Updated Name",
				},
			},
			wantErrCode: codes.NotFound,
			wantErr:     entity.ErrBookNotFound,
			mocksUsed:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.mocksUsed {
				bookUseCase.
					EXPECT().
					UpdateBook(ctx, test.args.req.GetId(), test.args.req.GetName(), test.args.req.GetAuthorId()).
					Return(test.wantErr)
			}

			got, err := service.UpdateBook(test.args.ctx, test.args.req)

			testutils.CheckError(t, err, test.wantErrCode)
			if err == nil {
				assert.NotNil(t, got)
				assert.IsType(t, &library.UpdateBookResponse{}, got)
			}
		})
	}
}
