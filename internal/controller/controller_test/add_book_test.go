package controller_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library/mocks"
	testutils "github.com/project/library/internal/usecase/library/test"
)

// На уровне контроллеров тестируется лишь валидация
// И корректное возвращение данных

// Тесты в пакете написаны с ошибкой, моки необходимо инициализировать для каждого сабтеста при использовании t.Parallel
// P.s. но работают корректно(удача)

// FIXME Необходимо перенести моки в сабтесты при использовании t.Parallel

func Test_AddBook(t *testing.T) {
	t.Parallel()                    // Разрешение на параллельный запуск тестов в рамках 1 пакета
	ctrl := gomock.NewController(t) // Управление жизненным циклом моков
	logger, _ := zap.NewProduction()
	authorUseCase := mocks.NewMockAuthorUseCase(ctrl)
	bookUseCase := mocks.NewMockBooksUseCase(ctrl)
	// Создаем grpc сервер с внедренными моками
	service := controller.New(logger, bookUseCase, authorUseCase)
	ctx := t.Context()

	type args struct {
		ctx context.Context
		req *library.AddBookRequest
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
			name: "add book | without authors",
			args: args{ctx,
				&library.AddBookRequest{
					Name:      "book1",
					AuthorIds: make([]string, 0),
				},
			},
			want: &entity.Book{
				ID:        uuid.NewString(),
				Name:      "book1",
				AuthorIDs: make([]string, 0),
			},
			wantErrCode: codes.OK,
			mocksUsed:   true,
		},

		{
			name: "add book | with authors",
			args: args{ctx,
				&library.AddBookRequest{
					Name:      "book2",
					AuthorIds: []string{"7a948d89-108c-4133-be30-788bd453c0cd"},
				},
			},
			want: &entity.Book{
				ID:        uuid.NewString(),
				Name:      "book2",
				AuthorIDs: []string{"7a948d89-108c-4133-be30-788bd453c0cd"},
			},
			wantErrCode: codes.OK,
			mocksUsed:   true,
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

			wantErrCode: codes.InvalidArgument,
			wantErr:     status.Error(codes.InvalidArgument, " invalid authors "),
			mocksUsed:   false,
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
			wantErrCode: codes.InvalidArgument,
			wantErr:     status.Error(codes.InvalidArgument, " invalid book name "),
			mocksUsed:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.mocksUsed {
				bookUseCase.
					// Описание действий заглушки
					EXPECT().
					RegisterBook(ctx, test.args.req.GetName(), test.args.req.GetAuthorIds()).
					Return(*test.want, test.wantErr)
			}

			got, err := service.AddBook(test.args.ctx, test.args.req)

			testutils.CheckError(t, err, test.wantErrCode)
			if err == nil && test.want != nil {
				assert.Equal(t, test.want.ID, got.GetBook().GetId())
				assert.Equal(t, test.want.Name, got.GetBook().GetName())
				assert.Equal(t, test.want.AuthorIDs, got.GetBook().GetAuthorId())
			}
		})
	}
}
