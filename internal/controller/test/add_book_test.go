package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/usecase/library/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// Проверка ожидаемой работы
// Ошибка: пришли не валидные данные
// Ошибка: ошибка с уровня usecase

// Возвращаемые ошибки не проверяются, я хз как это нормально сделать

func TestAddBook(t *testing.T) {
	t.Parallel() // Разрешение на параллельный запуск тестов в рамках 1 пакета
	logger, _ := zap.NewProduction()
	ctx := t.Context()

	type args struct {
		ctx context.Context
		req *library.AddBookRequest
	}

	tests := []struct {
		name      string
		args      args
		want      *library.AddBookResponse
		wantErr   bool
		mocksUsed bool
	}{
		{
			name: "add book | without authors",
			args: args{ctx,
				&library.AddBookRequest{
					Name:     "book1",
					AuthorId: make([]string, 0),
				},
			},
			want: &library.AddBookResponse{
				Book: &library.Book{
					Id:       uuid1,
					Name:     "book1",
					AuthorId: make([]string, 0),
				},
			},
			wantErr:   false,
			mocksUsed: true,
		},
		{
			name: "add book | with authors",
			args: args{ctx,
				&library.AddBookRequest{
					Name:     "book2",
					AuthorId: []string{uuid2, uuid3, uuid4},
				},
			},
			want: &library.AddBookResponse{
				Book: &library.Book{
					Id:       uuid5,
					Name:     "book2",
					AuthorId: []string{uuid2, uuid3, uuid4},
				},
			},
			wantErr:   false,
			mocksUsed: true,
		},
		{
			name: "add book | with invalid authors",
			args: args{
				ctx,
				&library.AddBookRequest{
					Name:     "book",
					AuthorId: []string{"1"},
				},
			},
			wantErr:   true,
			mocksUsed: false,
		},
		{
			name: "add book | with invalid name",
			args: args{
				ctx,
				&library.AddBookRequest{
					Name:     "",
					AuthorId: make([]string, 0),
				},
			},
			wantErr:   true,
			mocksUsed: false,
		},
		{
			name: "add book | usecase lvl error",
			args: args{
				ctx,
				&library.AddBookRequest{
					Name:     "book",
					AuthorId: make([]string, 0),
				},
			},
			wantErr:   true,
			mocksUsed: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel() // Параллельный запуск в цикле

			// Создаем моки внутри каждого сабтеста
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authorUseCase := mocks.NewMockAuthorUseCase(ctrl)
			bookUseCase := mocks.NewMockBooksUseCase(ctrl)

			// Создаем grpc сервер с внедренными моками
			service := controller.New(logger, bookUseCase, authorUseCase)

			if test.mocksUsed {
				var mockErr error
				if test.wantErr {
					mockErr = errors.New("mock error")
				}
				bookUseCase.
					// Описание действий заглушки
					EXPECT().
					AddBook(ctx, test.args.req.GetName(), test.args.req.GetAuthorId()).
					Return(test.want.Book, mockErr)
			}

			got, err := service.AddBook(test.args.ctx, test.args.req)

			if err == nil && test.want != nil {
				assert.Equal(t, test.want.GetBook().GetId(), got.GetBook().GetId())
				assert.Equal(t, test.want.GetBook().GetName(), got.GetBook().GetName())
				assert.Equal(t, test.want.GetBook().GetAuthorId(), got.GetBook().GetAuthorId())
			}

			if test.wantErr {
				assert.Error(t, err)
			}
		})
	}
}
