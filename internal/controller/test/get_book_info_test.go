package controller

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
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
		name      string
		args      args
		want      *library.GetBookInfoResponse
		wantErr   error
		mocksUsed bool
	}{
		{
			name: "get book info | valid request",
			args: args{
				ctx,
				&library.GetBookInfoRequest{
					Id: uuid6,
				},
			},
			want: &library.GetBookInfoResponse{
				Book: &library.Book{
					Id:        uuid6,
					Name:      "Test Book",
					AuthorIds: []string{uuid.NewString(), uuid.NewString()},
				},
			},
			wantErr:   nil,
			mocksUsed: true,
		},
		{
			name: "get book info | invalid uuid",
			args: args{
				ctx,
				&library.GetBookInfoRequest{
					Id: "abobus12",
				},
			},
			want:      nil,
			wantErr:   mockErr,
			mocksUsed: false,
		},
		{
			name: "get book info | empty id",
			args: args{
				ctx,
				&library.GetBookInfoRequest{
					Id: "",
				},
			},
			want:      nil,
			wantErr:   mockErr,
			mocksUsed: false,
		},
		{
			name: "get book info | book not found",
			args: args{
				ctx,
				&library.GetBookInfoRequest{
					Id: uuid6,
				},
			},
			want:      nil,
			wantErr:   mockErr,
			mocksUsed: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authorUseCase := mocks.NewMockAuthorUseCase(ctrl)
			bookUseCase := mocks.NewMockBooksUseCase(ctrl)
			service := controller.New(logger, bookUseCase, authorUseCase)

			if test.mocksUsed {
				var book *entity.Book
				if test.want != nil {
					book = ProtoToBook(test.want.Book)
				}

				bookUseCase.
					EXPECT().
					GetBook(gomock.Any(), test.args.req.GetId()).
					Return(book, test.wantErr)
			}

			got, err := service.GetBookInfo(test.args.ctx, test.args.req)

			if err == nil && test.want != nil {
				assert.Equal(t, test.want.Book.Id, got.GetBook().GetId())
				assert.Equal(t, test.want.Book.Name, got.GetBook().GetName())
				assert.Equal(t, test.want.Book.AuthorIds, got.GetBook().GetAuthorIds())
			}

			if test.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
