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

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library"
	"github.com/project/library/internal/usecase/repository"
	"github.com/project/library/internal/usecase/repository/mocks"
)

var defaultAuthor = &entity.Author{
	ID:   uuid.NewString(),
	Name: "name",
}

func TestRegisterAuthor(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	serialized, _ := json.Marshal(defaultAuthor)
	idempotencyKey := repository.OutboxKindAuthor.String() + "_" + defaultAuthor.ID

	tests := []struct {
		name                  string
		repositoryRerunAuthor *entity.Author
		returnAuthor          *entity.Author
		repositoryErr         error
		outboxErr             error
	}{
		{
			name:                  "register author",
			repositoryRerunAuthor: defaultAuthor,
			returnAuthor:          defaultAuthor,
			repositoryErr:         nil,
			outboxErr:             nil,
		},
		{
			name:                  "register author | repository error",
			repositoryRerunAuthor: nil,
			returnAuthor:          nil,
			repositoryErr:         errors.New("error register author"),
			outboxErr:             nil,
		},
		{
			name:                  "register author | outbox error",
			repositoryRerunAuthor: defaultAuthor,
			returnAuthor:          nil,
			repositoryErr:         nil,
			outboxErr:             errors.New("outbox error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mockAuthorRepo := mocks.NewMockAuthorRepository(ctrl)
			mockOutboxRepo := mocks.NewMockOutboxRepository(ctrl)
			mockTransactor := mocks.NewMockTransactor(ctrl)
			logger, _ := zap.NewProduction()
			useCase := library.New(logger, mockAuthorRepo,
				nil, mockOutboxRepo, mockTransactor)
			ctx := t.Context()

			mockTransactor.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(
				func(ctx context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				})

			mockAuthorRepo.EXPECT().RegisterAuthor(ctx, gomock.Any()).
				Return(test.repositoryRerunAuthor, test.repositoryErr)

			if test.repositoryErr == nil {
				mockOutboxRepo.EXPECT().SendMessage(ctx, idempotencyKey,
					repository.OutboxKindAuthor, serialized).Return(test.outboxErr)
			}
			var (
				resultAuthor *entity.Author
				err          error
			)

			if test.repositoryRerunAuthor == nil {
				resultAuthor, err = useCase.RegisterAuthor(ctx, defaultAuthor.Name)
			} else {
				resultAuthor, err = useCase.RegisterAuthor(ctx, test.repositoryRerunAuthor.Name)
			}

			switch {
			case test.outboxErr == nil && test.repositoryErr == nil:
				require.NoError(t, err)
			case test.outboxErr != nil:
				require.ErrorIs(t, err, test.outboxErr)
			case test.repositoryErr != nil:
				require.ErrorIs(t, err, test.repositoryErr)
			}

			assert.Equal(t, test.returnAuthor, resultAuthor)
		})
	}
}

func TestGetAuthorInfo(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	tests := []struct {
		name                  string
		repositoryRerunAuthor *entity.Author
		wantErr               error
		wantErrCode           codes.Code
	}{
		{
			name:                  "get author info",
			repositoryRerunAuthor: defaultAuthor,
		},
		{
			name:                  "get author info | with error",
			repositoryRerunAuthor: defaultAuthor,
			wantErr:               entity.ErrAuthorNotFound,
			wantErrCode:           codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mockAuthorRepo := mocks.NewMockAuthorRepository(ctrl)
			logger, _ := zap.NewProduction()
			useCase := library.New(logger, mockAuthorRepo,
				nil, nil, nil)
			ctx := t.Context()

			mockAuthorRepo.EXPECT().GetAuthorInfo(ctx, test.repositoryRerunAuthor.ID).Return(test.repositoryRerunAuthor, test.wantErr)

			got, wantErr := useCase.GetAuthorInfo(ctx, test.repositoryRerunAuthor.ID)
			CheckError(t, wantErr, test.wantErrCode)
			assert.Equal(t, test.repositoryRerunAuthor, got)
		})
	}
}

func TestChangeAuthor(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	tests := []struct {
		name                  string
		repositoryRerunAuthor *entity.Author
		wantErr               error
		wantErrCode           codes.Code
	}{
		{
			name:                  "change author",
			repositoryRerunAuthor: defaultAuthor,
		},
		{
			name:                  "change author | with error",
			repositoryRerunAuthor: defaultAuthor,
			wantErr:               entity.ErrAuthorNotFound,
			wantErrCode:           codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mockAuthorRepo := mocks.NewMockAuthorRepository(ctrl)
			logger, _ := zap.NewProduction()
			useCase := library.New(logger, mockAuthorRepo,
				nil, nil, nil)
			ctx := t.Context()

			mockAuthorRepo.EXPECT().ChangeAuthor(ctx, test.repositoryRerunAuthor.ID, test.repositoryRerunAuthor.Name).Return(test.wantErr)

			wantErr := useCase.ChangeAuthor(ctx, test.repositoryRerunAuthor.ID, test.repositoryRerunAuthor.Name)
			CheckError(t, wantErr, test.wantErrCode)
		})
	}
}
