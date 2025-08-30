package library

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library"
	"github.com/project/library/internal/usecase/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestRegisterAuthor(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	logger, _ := zap.NewProduction()

	tests := []struct {
		name        string
		authorName  string
		mockSetup   func(authorRepo *mocks.MockAuthorRepository)
		want        *entity.Author
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "register author usecase | successful author registration",
			authorName: "Aboba",
			mockSetup: func(authorRepo *mocks.MockAuthorRepository) {
				authorRepo.EXPECT().
					RegisterAuthor(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, author *entity.Author) (*entity.Author, error) {
						author.ID = uuid.NewString()
						assert.Equal(t, "Aboba", author.Name)
						assert.NotEmpty(t, author.ID)
						return author, nil
					})
			},
			want: &entity.Author{
				Name: "Aboba",
			},
			wantErr: false,
		},
		{
			name:       "register author usecase | repository error",
			authorName: "Test Author",
			mockSetup: func(authorRepo *mocks.MockAuthorRepository) {
				authorRepo.EXPECT().
					RegisterAuthor(ctx, gomock.Any()).
					Return(nil, errors.New("repository error"))
			},
			wantErr:     true,
			expectedErr: errors.New("repository error"),
		},
	}

	for _, test := range tests {
		test := test // capture range variable
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authorRepo := mocks.NewMockAuthorRepository(ctrl)
			bookRepo := mocks.NewMockBooksRepository(ctrl)
			repo := library.New(logger, authorRepo, bookRepo)

			if test.mockSetup != nil {
				test.mockSetup(authorRepo)
			}

			got, err := repo.RegisterAuthor(ctx, test.authorName)

			if test.wantErr {
				require.Error(t, err)
				if test.expectedErr != nil {
					assert.Equal(t, test.expectedErr.Error(), err.Error())
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want.Name, got.Name)
				assert.NotEmpty(t, got.ID)
			}
		})
	}
}

func TestGetAuthorInfo(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	logger, _ := zap.NewProduction()

	tests := []struct {
		name        string
		authorID    string
		mockSetup   func(authorRepo *mocks.MockAuthorRepository)
		want        *entity.Author
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "get author info usecase | successful get",
			authorID: "7a948d89-108c-4133-be30-788bd453c0cd",
			mockSetup: func(authorRepo *mocks.MockAuthorRepository) {
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, "7a948d89-108c-4133-be30-788bd453c0cd").
					Return(&entity.Author{
						ID:   "7a948d89-108c-4133-be30-788bd453c0cd",
						Name: "Author Name",
					}, nil)
			},
			want: &entity.Author{
				ID:   "7a948d89-108c-4133-be30-788bd453c0cd",
				Name: "Author Name",
			},
			wantErr: false,
		},
		{
			name:     "get author info usecase | repository error",
			authorID: "7a948d89-108c-4133-be30-788bd453c0cd",
			mockSetup: func(authorRepo *mocks.MockAuthorRepository) {
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, "7a948d89-108c-4133-be30-788bd453c0cd").
					Return(nil, errors.New("repository error"))
			},
			wantErr:     true,
			expectedErr: errors.New("repository error"),
		},
		{
			name:     "get author info usecase | author not found",
			authorID: "no-existent-id",
			mockSetup: func(authorRepo *mocks.MockAuthorRepository) {
				authorRepo.EXPECT().
					GetAuthorInfo(ctx, "no-existent-id").
					Return(nil, entity.ErrAuthorNotFound)
			},
			wantErr:     true,
			expectedErr: entity.ErrAuthorNotFound,
		},
	}

	for _, test := range tests {
		test := test // capture range variable
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authorRepo := mocks.NewMockAuthorRepository(ctrl)
			bookRepo := mocks.NewMockBooksRepository(ctrl)
			repo := library.New(logger, authorRepo, bookRepo)

			if test.mockSetup != nil {
				test.mockSetup(authorRepo)
			}

			got, err := repo.GetAuthorInfo(ctx, test.authorID)

			if test.wantErr {
				require.Error(t, err)
				if test.expectedErr != nil {
					assert.Equal(t, test.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want.ID, got.ID)
				assert.Equal(t, test.want.Name, got.Name)
			}
		})
	}
}

func TestChangeAuthor(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	logger, _ := zap.NewProduction()

	tests := []struct {
		name        string
		authorID    string
		authorName  string
		mockSetup   func(authorRepo *mocks.MockAuthorRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "update author usecase | successful update",
			authorID:   "7a948d89-108c-4133-be30-788bd453c0cd",
			authorName: "Updated Author Name",
			mockSetup: func(authorRepo *mocks.MockAuthorRepository) {
				authorRepo.EXPECT().
					ChangeAuthor(ctx, "7a948d89-108c-4133-be30-788bd453c0cd", "Updated Author Name").
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:       "update author usecase | repository error",
			authorID:   "7a948d89-108c-4133-be30-788bd453c0cd",
			authorName: "Updated Author Name",
			mockSetup: func(authorRepo *mocks.MockAuthorRepository) {
				authorRepo.EXPECT().
					ChangeAuthor(ctx, "7a948d89-108c-4133-be30-788bd453c0cd", "Updated Author Name").
					Return(errors.New("repository error"))
			},
			wantErr:     true,
			expectedErr: errors.New("repository error"),
		},
		{
			name:       "update author usecase | author not found",
			authorID:   "non-existent-id",
			authorName: "Updated Author Name",
			mockSetup: func(authorRepo *mocks.MockAuthorRepository) {
				authorRepo.EXPECT().
					ChangeAuthor(ctx, "non-existent-id", "Updated Author Name").
					Return(entity.ErrAuthorNotFound)
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
				test.mockSetup(authorRepo)
			}

			err := repo.ChangeAuthor(ctx, test.authorID, test.authorName)

			if test.wantErr {
				require.Error(t, err)
				if test.expectedErr != nil {
					assert.Equal(t, test.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
