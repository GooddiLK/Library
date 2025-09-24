package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/project/library/internal/usecase/repository"
)

func TestSendMassage(t *testing.T) {
	t.Parallel()

	idempotencyKey := "test-key"
	kind := repository.OutboxKindBook
	message := []byte("test-message")

	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "send massage",
			wantErr: nil,
		},
		{
			name:    "send massage | failure",
			wantErr: fmt.Errorf("test error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mockDB, err := pgxmock.NewPool()
			if err != nil {
				t.Fatal(err)
			}
			defer mockDB.Close()

			logger, _ := zap.NewProduction()
			outboxRepo := repository.NewOutbox(mockDB, logger)
			ctx := t.Context()

			if test.wantErr != nil {
				mockDB.ExpectExec("INSERT INTO outbox").
					WithArgs(idempotencyKey, message, kind).
					WillReturnError(test.wantErr)
			} else {
				mockDB.ExpectExec("INSERT INTO outbox").
					WithArgs(idempotencyKey, message, kind).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			}

			err = outboxRepo.SendMessage(ctx, idempotencyKey, kind, message)
			if test.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestGetMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		batchSize     int
		inProgressTTL time.Duration
		returnRows    *pgxmock.Rows
		wantErr       bool
		mockErr       error
		expectedData  []repository.OutboxData
	}{
		{
			name:          "get messages",
			batchSize:     2,
			inProgressTTL: 5 * time.Second,
			returnRows: pgxmock.NewRows([]string{"idempotency_key", "data", "kind"}).
				AddRow("key1", []byte("message1"), repository.OutboxKindBook).
				AddRow("key2", []byte("message2"), repository.OutboxKindBook),
			expectedData: []repository.OutboxData{
				{IdempotencyKey: "key1", RawData: []byte("message1"), Kind: repository.OutboxKindBook},
				{IdempotencyKey: "key2", RawData: []byte("message2"), Kind: repository.OutboxKindBook},
			},
			wantErr: false,
		},
		{
			name:          "get messages | query error",
			batchSize:     2,
			inProgressTTL: 5 * time.Second,
			returnRows:    nil,
			wantErr:       true,
			mockErr:       fmt.Errorf("test error"),
			expectedData:  nil,
		},
		{
			name:          "get messages | scan error",
			batchSize:     2,
			inProgressTTL: 5 * time.Second,
			returnRows: pgxmock.NewRows([]string{"idempotency_key", "data", "kind"}).
				AddRow("key1", []byte("message1"), repository.OutboxKindBook).
				AddRow("key2", nil, "1"),
			expectedData: nil,
			wantErr:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mockDB, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockDB.Close()

			logger, _ := zap.NewProduction()
			outboxRepo := repository.NewOutbox(mockDB, logger)
			ctx := t.Context()

			interval := fmt.Sprintf("%d ms", test.inProgressTTL.Milliseconds())
			if test.mockErr != nil {
				mockDB.ExpectQuery("UPDATE outbox").
					WithArgs(interval, test.batchSize).
					WillReturnError(test.mockErr)
			} else {
				mockDB.ExpectQuery("UPDATE outbox").
					WithArgs(interval, test.batchSize).
					WillReturnRows(test.returnRows)
			}

			data, err := outboxRepo.GetMessages(ctx, test.batchSize, test.inProgressTTL)

			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedData, data)
			}

			require.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestMarkAsProcessed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		idempotencyKeys []string
		mockErr         error
		wantErr         bool
		returnResult    pgconn.CommandTag
	}{
		{
			name:            "mark as processed successfully",
			idempotencyKeys: []string{"key1", "key2"},
			mockErr:         nil,
			wantErr:         false,
		},

		{
			name:            "database error",
			idempotencyKeys: []string{"key1", "key2"},
			mockErr:         fmt.Errorf("database error"),
			wantErr:         true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mockDB, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockDB.Close()

			logger, _ := zap.NewProduction()
			outboxRepo := repository.NewOutbox(mockDB, logger)
			ctx := t.Context()

			if test.mockErr != nil {
				mockDB.ExpectExec("UPDATE outbox").
					WithArgs(test.idempotencyKeys).
					WillReturnError(test.mockErr)
			} else {
				mockDB.ExpectExec("UPDATE outbox").
					WithArgs(test.idempotencyKeys).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			}

			err = outboxRepo.MarkAsProcessed(ctx, test.idempotencyKeys)

			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}
