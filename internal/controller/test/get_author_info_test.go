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

func Test_GetAuthorInfo(t *testing.T) {
	t.Parallel()
	logger, _ := zap.NewProduction()
	ctx := t.Context()

	type args struct {
		ctx context.Context
		req *library.GetAuthorInfoRequest
	}

	tests := []struct {
		name      string
		args      args
		want      *library.GetAuthorInfoResponse
		wantErr   error
		mocksUsed bool
	}{
		{
			name: "get author info | valid request",
			args: args{
				ctx,
				&library.GetAuthorInfoRequest{
					Id: uuid7,
				},
			},
			want: &library.GetAuthorInfoResponse{
				Id:   uuid7,
				Name: "Author Name",
			},
			wantErr:   nil,
			mocksUsed: true,
		},
		{
			name: "get author info | invalid id",
			args: args{
				ctx,
				&library.GetAuthorInfoRequest{
					Id: "1aboba2",
				},
			},
			want:      nil,
			wantErr:   mockErr,
			mocksUsed: false,
		},
		{
			name: "get author info | empty id",
			args: args{
				ctx,
				&library.GetAuthorInfoRequest{
					Id: "",
				},
			},
			want:      nil,
			wantErr:   mockErr,
			mocksUsed: false,
		},
		{
			name: "get author info | author not found",
			args: args{
				ctx,
				&library.GetAuthorInfoRequest{
					Id: uuid.NewString(),
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
				var auth *entity.Author
				if test.want != nil {
					auth = &entity.Author{Id: test.want.Id, Name: test.want.Name}
				}

				authorUseCase.
					EXPECT().
					GetAuthorInfo(gomock.Any(), test.args.req.GetId()).
					Return(auth, test.wantErr)
			}

			got, err := service.GetAuthorInfo(test.args.ctx, test.args.req)

			if err == nil && test.want != nil {
				assert.Equal(t, test.want.GetId(), got.GetId())
				assert.Equal(t, test.want.GetName(), got.GetName())
			}

			if test.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
