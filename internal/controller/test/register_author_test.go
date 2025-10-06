package controller

import (
	"context"
	"strings"
	"testing"

	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// Тесту проверяющему наличие автора стоило бы существовать, но два автора с одним именем могут существовать
func Test_RegisterAuthor(t *testing.T) {
	t.Parallel()
	logger, _ := zap.NewProduction()
	ctx := t.Context()

	type args struct {
		ctx context.Context
		req *library.RegisterAuthorRequest
	}

	tests := []struct {
		name      string
		args      args
		want      *library.RegisterAuthorResponse
		wantErr   error
		mocksUsed bool
	}{
		{
			name: "register author | valid request",
			args: args{
				ctx,
				&library.RegisterAuthorRequest{
					Name: "NameAbobs",
				},
			},
			want: &library.RegisterAuthorResponse{
				Id: uuid1,
			},
			wantErr:   nil,
			mocksUsed: true,
		},
		{
			name: "register author | empty name",
			args: args{
				ctx,
				&library.RegisterAuthorRequest{
					Name: "",
				},
			},
			want:      nil,
			wantErr:   mockErr,
			mocksUsed: false,
		},
		{
			name: "register author | usecase error",
			args: args{
				ctx,
				&library.RegisterAuthorRequest{
					Name: strings.Repeat("A", 511),
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
					auth = &entity.Author{Id: test.want.Id}
				}

				authorUseCase.
					EXPECT().
					RegisterAuthor(gomock.Any(), test.args.req.GetName()).
					Return(auth, test.wantErr)
			}

			got, err := service.RegisterAuthor(test.args.ctx, test.args.req)

			if err == nil && test.want != nil {
				assert.Equal(t, test.want.Id, got.GetId())
			}

			if test.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
