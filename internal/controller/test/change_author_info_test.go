package controller

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/google/uuid"
	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/usecase/library/mocks"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// Проверка ожидаемой работы
// Ошибка: пришли не валидные данные
// Ошибка: ошибка с уровня usecase

func Test_ChangeAuthorInfo(t *testing.T) {
	t.Parallel()
	logger, _ := zap.NewProduction()
	ctx := t.Context()

	type args struct {
		ctx context.Context
		req *library.ChangeAuthorInfoRequest
	}

	tests := []struct {
		name      string
		args      args
		wantErr   error
		mocksUsed bool
	}{
		{
			"change author info | without error",
			args{ctx,
				&library.ChangeAuthorInfoRequest{
					Id:   uuid.NewString(),
					Name: "New Name",
				},
			},
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
			mockErr,
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
			mockErr,
			false,
		},
		{
			"change author info | usecase error",
			args{
				ctx,
				&library.ChangeAuthorInfoRequest{
					Id:   uuid.NewString(),
					Name: "New Name",
				},
			},
			mockErr,
			true,
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
				authorUseCase.EXPECT().
					ChangeAuthor(gomock.Any(), test.args.req.GetId(), test.args.req.GetName()).
					Return(test.wantErr)
			}

			got, err := service.ChangeAuthorInfo(test.args.ctx, test.args.req)

			if test.wantErr == nil {
				assert.NotNil(t, got)
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
