package library

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library"
	"github.com/project/library/internal/usecase/repository"
	"github.com/project/library/internal/usecase/repository/mocks"
)

func TestAddBook(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	book := &entity.Book{
		Id:        "book-id",
		Name:      "Test Book",
		AuthorIds: []string{"author1", "author2"},
	}
	serialized, _ := json.Marshal(book)
	idempotencyKey := repository.OutboxKindBook.String() + "_" + book.Id

	tests := []struct {
		name                string
		repositoryRerunBook *entity.Book
		returnBook          *entity.Book
		repositoryErr       error
		outboxErr           error
	}{
		{
			name:                "add book",
			repositoryRerunBook: book,
			returnBook:          book,
			repositoryErr:       nil,
			outboxErr:           nil,
		},
		{
			name:                "add book | repository error",
			repositoryRerunBook: nil,
			returnBook:          nil,
			repositoryErr:       entity.ErrAuthorNotFound,
			outboxErr:           nil,
		},
		{
			name:                "add book | outbox error",
			repositoryRerunBook: book,
			returnBook:          nil,
			repositoryErr:       nil,
			outboxErr:           errors.New("cannot send message"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mockBooksRepo := mocks.NewMockBooksRepository(ctrl)
			mockOutboxRepo := mocks.NewMockOutboxRepository(ctrl)
			mockTransactor := mocks.NewMockTransactor(ctrl)
			logger, _ := zap.NewProduction()
			useCase := library.New(logger, nil,
				mockBooksRepo, mockOutboxRepo, mockTransactor)
			ctx := t.Context()

			mockTransactor.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(
				func(ctx context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				},
			)
			mockBooksRepo.EXPECT().AddBook(ctx, gomock.Any()).
				Return(test.repositoryRerunBook, test.repositoryErr)

			if test.repositoryErr == nil {
				mockOutboxRepo.EXPECT().SendMessage(ctx, idempotencyKey,
					repository.OutboxKindBook, serialized).Return(test.outboxErr)
			}

			resultBook, err := useCase.AddBook(ctx, book.Name, book.AuthorIds)
			switch {
			case test.outboxErr == nil && test.repositoryErr == nil:
				require.NoError(t, err)
			case test.outboxErr != nil:
				require.ErrorIs(t, err, test.outboxErr)
			case test.repositoryErr != nil:
				require.ErrorIs(t, err, test.repositoryErr)
			}

			assert.Equal(t, test.returnBook, resultBook)
		})
	}
}

func TestGetBook(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	tests := []struct {
		name        string
		returnBook  *entity.Book
		wantErr     error
		wantErrCode codes.Code
	}{
		{
			name: "get book",
			returnBook: &entity.Book{
				Id:        uuid.NewString(),
				Name:      "name",
				AuthorIds: make([]string, 0),
			},
		},
		{
			name: "get book | with error",
			returnBook: &entity.Book{
				Id:        uuid.NewString(),
				Name:      "name",
				AuthorIds: make([]string, 0),
			},
			wantErrCode: codes.Internal,
			wantErr:     status.Error(codes.Internal, "error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mockBookRepo := mocks.NewMockBooksRepository(ctrl)
			logger, _ := zap.NewProduction()
			useCase := library.New(logger, nil,
				mockBookRepo, nil, nil)
			ctx := t.Context()

			mockBookRepo.EXPECT().GetBook(ctx, test.returnBook.Id).
				Return(test.returnBook, test.wantErr)

			got, err := useCase.GetBook(ctx, test.returnBook.Id)
			CheckError(t, err, test.wantErrCode)
			assert.Equal(t, test.returnBook, got)
		})
	}
}

func TestUpdateBook(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	tests := []struct {
		name        string
		returnBook  *entity.Book
		wantErr     error
		wantErrCode codes.Code
	}{
		{
			name: "update book",
			returnBook: &entity.Book{
				Id:        uuid.NewString(),
				Name:      "name",
				AuthorIds: make([]string, 0),
			},
		},
		{
			name: "update book | with error",
			returnBook: &entity.Book{
				Id:        uuid.NewString(),
				Name:      "name",
				AuthorIds: make([]string, 0),
			},
			wantErrCode: codes.NotFound,
			wantErr:     entity.ErrBookNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mockBookRepo := mocks.NewMockBooksRepository(ctrl)
			logger, _ := zap.NewProduction()
			useCase := library.New(logger, nil,
				mockBookRepo, nil, nil)
			ctx := t.Context()

			mockBookRepo.EXPECT().UpdateBook(ctx,
				test.returnBook.Id, test.returnBook.Name, test.returnBook.AuthorIds).
				Return(test.wantErr)

			err := useCase.UpdateBook(ctx,
				test.returnBook.Id, test.returnBook.Name, test.returnBook.AuthorIds)
			CheckError(t, err, test.wantErrCode)
		})
	}
}

func TestGetAuthorBooks(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	tests := []struct {
		name                  string
		repositoryRerunAuthor *entity.Author
		returnBooks           []*entity.Book
		wantErr               error
		wantErrCode           codes.Code
	}{
		{
			name:                  "get author books",
			repositoryRerunAuthor: defaultAuthor,
			wantErr:               entity.ErrAuthorNotFound,
			wantErrCode:           codes.NotFound,
		},
		{
			name:                  "get author books | with error",
			repositoryRerunAuthor: defaultAuthor,
			returnBooks: []*entity.Book{
				{Name: "first book"},
				{Name: "second book"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mockBooksRepo := mocks.NewMockBooksRepository(ctrl)
			logger, _ := zap.NewProduction()
			useCase := library.New(logger, nil,
				mockBooksRepo, nil, nil)
			ctx := t.Context()

			mockBooksRepo.EXPECT().GetAuthorBooks(ctx, test.repositoryRerunAuthor.Id).Return(test.returnBooks, test.wantErr)

			books, wantErr := useCase.GetAuthorBooks(ctx, test.repositoryRerunAuthor.Id)
			CheckError(t, wantErr, test.wantErrCode)
			assert.Equal(t, test.returnBooks, books)
		})
	}
}
