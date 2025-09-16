package library

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library"
	"github.com/project/library/internal/usecase/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestLibraryImpl_RegisterBook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	logger, _ := zap.NewProduction()

	tests := []struct {
		name        string
		bookName    string
		authorIDs   []string
		mockSetup   func(authorRepo *mocks.MockAuthorRepository, bookRepo *mocks.MockBooksRepository)
		want        entity.Book
		wantErr     bool
		expectedErr error
	}{
		{
			name:      "register book usecase | successful registration without authors",
			bookName:  "Test Book",
			authorIDs: []string{},
			mockSetup: func(authorRepo *mocks.MockAuthorRepository, bookRepo *mocks.MockBooksRepository) {
				bookRepo.EXPECT().
					CreateBook(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, book entity.Book) (entity.Book, error) {
						assert.Equal(t, "Test Book", book.Name)
						assert.Equal(t, []string{}, book.AuthorIDs)
						assert.NotEmpty(t, book.ID)
						return book, nil
					})
			},
			want: entity.Book{
				Name:      "Test Book",
				AuthorIDs: []string{},
			},
			wantErr: false,
		},
		{
			name:      "register book usecase | successful registration with authors",
			bookName:  "Test Book",
			authorIDs: []string{"550e8400-e29b-41d4-a716-446655440000", "1b9d6bcd-bbfd-4b2d-9b5d-ab8dfbbd4bed"},
			mockSetup: func(authorRepo *mocks.MockAuthorRepository, bookRepo *mocks.MockBooksRepository) {
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return("Author 1", nil)
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, "1b9d6bcd-bbfd-4b2d-9b5d-ab8dfbbd4bed").
					Return("Author 2", nil)

				bookRepo.EXPECT().
					CreateBook(ctx, gomock.Any()).
					Return(entity.Book{
						ID:        "7a948d89-108c-4133-be30-788bd453c0cd",
						Name:      "Test Book",
						AuthorIDs: []string{"550e8400-e29b-41d4-a716-446655440000", "1b9d6bcd-bbfd-4b2d-9b5d-ab8dfbbd4bed"},
					}, nil)
			},
			want: entity.Book{
				ID:        "7a948d89-108c-4133-be30-788bd453c0cd",
				Name:      "Test Book",
				AuthorIDs: []string{"550e8400-e29b-41d4-a716-446655440000", "1b9d6bcd-bbfd-4b2d-9b5d-ab8dfbbd4bed"},
			},
			wantErr: false,
		},
		{
			name:      "register book usecase | repository error",
			bookName:  "Test Book",
			authorIDs: []string{uuid.NewString()},
			mockSetup: func(authorRepo *mocks.MockAuthorRepository, bookRepo *mocks.MockBooksRepository) {
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, gomock.Any()).
					Return("Author Name", nil)

				bookRepo.EXPECT().
					CreateBook(ctx, gomock.Any()).
					Return(entity.Book{}, errors.New("repository error"))
			},
			wantErr:     true,
			expectedErr: errors.New("repository error"),
		},
		{
			name:      "register book usecase | author not found error",
			bookName:  "Test Book",
			authorIDs: []string{"non-existent-author"},
			mockSetup: func(authorRepo *mocks.MockAuthorRepository, bookRepo *mocks.MockBooksRepository) {
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, "non-existent-author").
					Return("", entity.ErrAuthorNotFound)
				// bookRepo.CreateBook НЕ должен вызываться
			},
			wantErr:     true,
			expectedErr: entity.ErrAuthorNotFound,
		},
		{
			name:      "register book usecase | one author not found among multiple",
			bookName:  "Test Book",
			authorIDs: []string{"existing-author", "non-existent-author"},
			mockSetup: func(authorRepo *mocks.MockAuthorRepository, bookRepo *mocks.MockBooksRepository) {
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, "existing-author").
					Return("Existing Author", nil)
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, "non-existent-author").
					Return("", entity.ErrAuthorNotFound)
			},
			wantErr:     true,
			expectedErr: entity.ErrAuthorNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authorRepo := mocks.NewMockAuthorRepository(ctrl)
			bookRepo := mocks.NewMockBooksRepository(ctrl)
			repo := library.New(logger, authorRepo, bookRepo)

			if test.mockSetup != nil {
				test.mockSetup(authorRepo, bookRepo)
			}

			got, err := repo.RegisterBook(ctx, test.bookName, test.authorIDs)

			if test.wantErr {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.want.Name, got.Name)
				assert.Equal(t, test.want.AuthorIDs, got.AuthorIDs)
				assert.NotEmpty(t, got.ID)
			}
		})
	}
}

func TestLibraryImpl_GetBook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	logger, _ := zap.NewProduction()

	tests := []struct {
		name        string
		bookID      string
		mockSetup   func(bookRepo *mocks.MockBooksRepository)
		want        entity.Book
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "get book usecase | successful get",
			bookID: "7a948d89-108c-4133-be30-788bd453c0cd",
			mockSetup: func(bookRepo *mocks.MockBooksRepository) {
				bookRepo.EXPECT().
					GetBook(ctx, "7a948d89-108c-4133-be30-788bd453c0cd").
					Return(entity.Book{
						ID:        "7a948d89-108c-4133-be30-788bd453c0cd",
						Name:      "Test Book",
						AuthorIDs: []string{"550e8400-e29b-41d4-a716-446655440000"},
					}, nil)
			},
			want: entity.Book{
				ID:        "7a948d89-108c-4133-be30-788bd453c0cd",
				Name:      "Test Book",
				AuthorIDs: []string{"550e8400-e29b-41d4-a716-446655440000"},
			},
			wantErr: false,
		},
		{
			name:   "get book usecase | repository error",
			bookID: "7a948d89-108c-4133-be30-788bd453c0cd",
			mockSetup: func(bookRepo *mocks.MockBooksRepository) {
				bookRepo.EXPECT().
					GetBook(ctx, "7a948d89-108c-4133-be30-788bd453c0cd").
					Return(entity.Book{}, errors.New("repository error"))
			},
			wantErr:     true,
			expectedErr: errors.New("repository error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authorRepo := mocks.NewMockAuthorRepository(ctrl)
			bookRepo := mocks.NewMockBooksRepository(ctrl)
			repo := library.New(logger, authorRepo, bookRepo)

			if test.mockSetup != nil {
				test.mockSetup(bookRepo)
			}

			got, err := repo.GetBook(ctx, test.bookID)

			if test.wantErr {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.want.ID, got.ID)
				assert.Equal(t, test.want.Name, got.Name)
				assert.Equal(t, test.want.AuthorIDs, got.AuthorIDs)
			}
		})
	}
}

func TestLibraryImpl_UpdateBook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	logger, _ := zap.NewProduction()

	tests := []struct {
		name        string
		bookID      string
		bookName    string
		authorIDs   []string
		mockSetup   func(authorRepo *mocks.MockAuthorRepository, bookRepo *mocks.MockBooksRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:      "update book usecase | successful update with name and authors",
			bookID:    "7a948d89-108c-4133-be30-788bd453c0cd",
			bookName:  "Updated Book Name",
			authorIDs: []string{"550e8400-e29b-41d4-a716-446655440000", "1b9d6bcd-bbfd-4b2d-9b5d-ab8dfbbd4bed"},
			mockSetup: func(authorRepo *mocks.MockAuthorRepository, bookRepo *mocks.MockBooksRepository) {
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return("Author 1", nil)
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, "1b9d6bcd-bbfd-4b2d-9b5d-ab8dfbbd4bed").
					Return("Author 2", nil)

				bookRepo.EXPECT().
					UpdateBook(ctx, entity.Book{
						ID:        "7a948d89-108c-4133-be30-788bd453c0cd",
						Name:      "Updated Book Name",
						AuthorIDs: []string{"550e8400-e29b-41d4-a716-446655440000", "1b9d6bcd-bbfd-4b2d-9b5d-ab8dfbbd4bed"},
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "update book usecase | successful update with name only",
			bookID:    "7a948d89-108c-4133-be30-788bd453c0cd",
			bookName:  "Updated Book Name",
			authorIDs: []string{},
			mockSetup: func(authorRepo *mocks.MockAuthorRepository, bookRepo *mocks.MockBooksRepository) {
				bookRepo.EXPECT().
					UpdateBook(ctx, entity.Book{
						ID:        "7a948d89-108c-4133-be30-788bd453c0cd",
						Name:      "Updated Book Name",
						AuthorIDs: []string{},
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "update book usecase | successful update with authors only",
			bookID:    "7a948d89-108c-4133-be30-788bd453c0cd",
			bookName:  "",
			authorIDs: []string{"550e8400-e29b-41d4-a716-446655440000"},
			mockSetup: func(authorRepo *mocks.MockAuthorRepository, bookRepo *mocks.MockBooksRepository) {
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return("Author Name", nil)

				bookRepo.EXPECT().
					UpdateBook(ctx, entity.Book{
						ID:        "7a948d89-108c-4133-be30-788bd453c0cd",
						Name:      "",
						AuthorIDs: []string{"550e8400-e29b-41d4-a716-446655440000"},
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "update book usecase | repository error",
			bookID:    "7a948d89-108c-4133-be30-788bd453c0cd",
			bookName:  "Updated Book Name",
			authorIDs: []string{"author-id-1"},
			mockSetup: func(authorRepo *mocks.MockAuthorRepository, bookRepo *mocks.MockBooksRepository) {
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, "author-id-1").
					Return("Author Name", nil)

				bookRepo.EXPECT().
					UpdateBook(ctx, gomock.Any()).
					Return(errors.New("repository error"))
			},
			wantErr:     true,
			expectedErr: errors.New("repository error"),
		},
		{
			name:      "update book usecase | author not found",
			bookID:    "7a948d89-108c-4133-be30-788bd453c0cd",
			bookName:  "Updated Book Name",
			authorIDs: []string{"non-existent-author"},
			mockSetup: func(authorRepo *mocks.MockAuthorRepository, bookRepo *mocks.MockBooksRepository) {
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, "non-existent-author").
					Return("", entity.ErrAuthorNotFound)
			},
			wantErr:     true,
			expectedErr: entity.ErrAuthorNotFound,
		},
		{
			name:      "update book usecase | one author not found among multiple",
			bookID:    "7a948d89-108c-4133-be30-788bd453c0cd",
			bookName:  "Updated Book Name",
			authorIDs: []string{"existing-author", "non-existent-author"},
			mockSetup: func(authorRepo *mocks.MockAuthorRepository, bookRepo *mocks.MockBooksRepository) {
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, "existing-author").
					Return("Existing Author", nil)
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, "non-existent-author").
					Return("", entity.ErrAuthorNotFound)
			},
			wantErr:     true,
			expectedErr: entity.ErrAuthorNotFound,
		},
		{
			name:      "update book usecase | book not found",
			bookID:    "non-existent-id",
			bookName:  "Updated Book Name",
			authorIDs: []string{"author-id-1"},
			mockSetup: func(authorRepo *mocks.MockAuthorRepository, bookRepo *mocks.MockBooksRepository) {
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, "author-id-1").
					Return("Author Name", nil)
				bookRepo.EXPECT().
					UpdateBook(ctx, gomock.Any()).
					Return(entity.ErrBookNotFound)
			},
			wantErr:     true,
			expectedErr: entity.ErrBookNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authorRepo := mocks.NewMockAuthorRepository(ctrl)
			bookRepo := mocks.NewMockBooksRepository(ctrl)
			repo := library.New(logger, authorRepo, bookRepo)

			if test.mockSetup != nil {
				test.mockSetup(authorRepo, bookRepo)
			}

			err := repo.UpdateBook(ctx, test.bookID, test.bookName, test.authorIDs)

			if test.wantErr {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLibraryImpl_GetAuthorBooks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	logger, _ := zap.NewProduction()

	tests := []struct {
		name        string
		authorID    string
		mockSetup   func(bookRepo *mocks.MockBooksRepository)
		want        []entity.Book
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "get author books usecase | successful get with multiple books",
			authorID: "550e8400-e29b-41d4-a716-446655440000",
			mockSetup: func(bookRepo *mocks.MockBooksRepository) {
				bookRepo.EXPECT().
					GetAuthorBooks(ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return([]entity.Book{
						{
							ID:        "book-id-1",
							Name:      "Book 1",
							AuthorIDs: []string{"550e8400-e29b-41d4-a716-446655440000"},
						},
						{
							ID:        "book-id-2",
							Name:      "Book 2",
							AuthorIDs: []string{"550e8400-e29b-41d4-a716-446655440000"},
						},
					}, nil)
			},
			want: []entity.Book{
				{
					ID:        "book-id-1",
					Name:      "Book 1",
					AuthorIDs: []string{"550e8400-e29b-41d4-a716-446655440000"},
				},
				{
					ID:        "book-id-2",
					Name:      "Book 2",
					AuthorIDs: []string{"550e8400-e29b-41d4-a716-446655440000"},
				},
			},
			wantErr: false,
		},
		{
			name:     "get author books usecase | successful get with single book",
			authorID: "1b9d6bcd-bbfd-4b2d-9b5d-ab8dfbbd4bed",
			mockSetup: func(bookRepo *mocks.MockBooksRepository) {
				bookRepo.EXPECT().
					GetAuthorBooks(ctx, "1b9d6bcd-bbfd-4b2d-9b5d-ab8dfbbd4bed").
					Return([]entity.Book{
						{
							ID:        "single-book-id",
							Name:      "Single Book",
							AuthorIDs: []string{"1b9d6bcd-bbfd-4b2d-9b5d-ab8dfbbd4bed"},
						},
					}, nil)
			},
			want: []entity.Book{
				{
					ID:        "single-book-id",
					Name:      "Single Book",
					AuthorIDs: []string{"1b9d6bcd-bbfd-4b2d-9b5d-ab8dfbbd4bed"},
				},
			},
			wantErr: false,
		},
		{
			name:     "get author books usecase | successful get with no books",
			authorID: "author-with-no-books",
			mockSetup: func(bookRepo *mocks.MockBooksRepository) {
				bookRepo.EXPECT().
					GetAuthorBooks(ctx, "author-with-no-books").
					Return([]entity.Book{}, nil)
			},
			want:    []entity.Book{},
			wantErr: false,
		},
		{
			name:     "get author books usecase | repository error",
			authorID: "7a948d89-108c-4133-be30-788bd453c0cd",
			mockSetup: func(bookRepo *mocks.MockBooksRepository) {
				bookRepo.EXPECT().
					GetAuthorBooks(ctx, "7a948d89-108c-4133-be30-788bd453c0cd").
					Return(nil, errors.New("repository error"))
			},
			wantErr:     true,
			expectedErr: errors.New("repository error"),
		},
		{
			name:     "get author books usecase | books with multiple authors",
			authorID: "multi-author-book-author",
			mockSetup: func(bookRepo *mocks.MockBooksRepository) {
				bookRepo.EXPECT().
					GetAuthorBooks(ctx, "multi-author-book-author").
					Return([]entity.Book{
						{
							ID:        "multi-author-book-1",
							Name:      "Multi Author Book 1",
							AuthorIDs: []string{"multi-author-book-author", "other-author-1"},
						},
						{
							ID:        "multi-author-book-2",
							Name:      "Multi Author Book 2",
							AuthorIDs: []string{"multi-author-book-author", "other-author-2"},
						},
					}, nil)
			},
			want: []entity.Book{
				{
					ID:        "multi-author-book-1",
					Name:      "Multi Author Book 1",
					AuthorIDs: []string{"multi-author-book-author", "other-author-1"},
				},
				{
					ID:        "multi-author-book-2",
					Name:      "Multi Author Book 2",
					AuthorIDs: []string{"multi-author-book-author", "other-author-2"},
				},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authorRepo := mocks.NewMockAuthorRepository(ctrl)
			bookRepo := mocks.NewMockBooksRepository(ctrl)
			repo := library.New(logger, authorRepo, bookRepo)

			if test.mockSetup != nil {
				test.mockSetup(bookRepo)
			}

			got, err := repo.GetAuthorBooks(ctx, test.authorID)

			if test.wantErr {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(test.want), len(got))

				for i, expectedBook := range test.want {
					assert.Equal(t, expectedBook.ID, got[i].ID)
					assert.Equal(t, expectedBook.Name, got[i].Name)
					assert.Equal(t, expectedBook.AuthorIDs, got[i].AuthorIDs)
				}
			}
		})
	}
}
