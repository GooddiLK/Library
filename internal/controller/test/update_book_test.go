package controller

import (
	"context"
	"testing"

	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/usecase/library/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func Test_UpdateBook(t *testing.T) {
	t.Parallel()
	logger, _ := zap.NewProduction()
	ctx := t.Context()

	type args struct {
		ctx context.Context
		req *library.UpdateBookRequest
	}

	tests := []struct {
		name      string
		args      args
		wantErr   error
		mocksUsed bool
	}{
		{
			name: "update book | valid request with name and authors",
			args: args{
				ctx,
				&library.UpdateBookRequest{
					Id:        uuid4,
					Name:      "New name",
					AuthorIds: []string{uuid5, uuid6},
				},
			},
			wantErr:   nil,
			mocksUsed: true,
		},
		{
			name: "update book | valid request with name only",
			args: args{
				ctx,
				&library.UpdateBookRequest{
					Id:   uuid8,
					Name: "New name",
				},
			},
			wantErr:   nil,
			mocksUsed: true,
		},
		{
			name: "update book | invalid request with authors only",
			args: args{
				ctx,
				&library.UpdateBookRequest{
					Id:        uuid4,
					AuthorIds: []string{"author-id-1"},
				},
			},
			wantErr:   mockErr,
			mocksUsed: false,
		},
		{
			name: "update book | invalid uuid",
			args: args{
				ctx,
				&library.UpdateBookRequest{
					Id: "aboba",
				},
			},
			wantErr:   mockErr,
			mocksUsed: false,
		},
		{
			name: "update book | empty id",
			args: args{
				ctx,
				&library.UpdateBookRequest{
					Id: "",
				},
			},
			wantErr:   mockErr,
			mocksUsed: false,
		},
		{
			name: "update book | book not found",
			args: args{
				ctx,
				&library.UpdateBookRequest{
					Id:   uuid4,
					Name: "Updated Name",
				},
			},
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
				bookUseCase.
					EXPECT().
					UpdateBook(gomock.Any(), test.args.req.GetId(), test.args.req.GetName(), test.args.req.GetAuthorIds()).
					Return(test.wantErr)
			}

			got, err := service.UpdateBook(test.args.ctx, test.args.req)

			if test.wantErr == nil {
				assert.NotNil(t, got)
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
