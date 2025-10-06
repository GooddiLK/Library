package controller

import (
	"context"
	"github.com/project/library/internal/entity"
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
		wantErr   error
		mocksUsed bool
	}{
		{
			name: "add book | without authors",
			args: args{ctx,
				&library.AddBookRequest{
					Name:      "book1",
					AuthorIds: make([]string, 0),
				},
			},
			want: &library.AddBookResponse{
				Book: &library.Book{
					Id:        uuid1,
					Name:      "book1",
					AuthorIds: make([]string, 0),
				},
			},
			wantErr:   nil,
			mocksUsed: true,
		},
		{
			name: "add book | with authors",
			args: args{ctx,
				&library.AddBookRequest{
					Name:      "book2",
					AuthorIds: []string{uuid2, uuid3, uuid4},
				},
			},
			want: &library.AddBookResponse{
				Book: &library.Book{
					Id:        uuid5,
					Name:      "book2",
					AuthorIds: []string{uuid2, uuid3, uuid4},
				},
			},
			wantErr:   nil,
			mocksUsed: true,
		},
		{
			name: "add book | with invalid authors",
			args: args{
				ctx,
				&library.AddBookRequest{
					Name:      "book",
					AuthorIds: []string{"1"},
				},
			},
			wantErr:   mockErr,
			mocksUsed: false,
		},
		{
			name: "add book | with invalid name",
			args: args{
				ctx,
				&library.AddBookRequest{
					Name:      "",
					AuthorIds: make([]string, 0),
				},
			},
			wantErr:   mockErr,
			mocksUsed: false,
		},
		{
			name: "add book | usecase lvl error",
			args: args{
				ctx,
				&library.AddBookRequest{
					Name:      "book",
					AuthorIds: make([]string, 0),
				},
			},
			want:      nil,
			wantErr:   mockErr,
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
			service := controller.New(logger, bookUseCase, authorUseCase)

			if test.mocksUsed {
				// Описание действий заглушки
				var book *entity.Book
				if test.want != nil {
					book = ProtoToBook(test.want.Book)
				}

				bookUseCase.
					EXPECT().
					AddBook(gomock.Any(), test.args.req.GetName(), test.args.req.GetAuthorIds()).
					Return(book, test.wantErr)
			}

			got, err := service.AddBook(test.args.ctx, test.args.req)

			if err == nil && test.want != nil {
				assert.Equal(t, test.want.GetBook().GetId(), got.GetBook().GetId())
				assert.Equal(t, test.want.GetBook().GetName(), got.GetBook().GetName())
				assert.Equal(t, test.want.GetBook().GetAuthorIds(), got.GetBook().GetAuthorIds())
			}

			if test.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
