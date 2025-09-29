package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var _ OutboxRepository = (*outboxRepository)(nil)

type outboxRepository struct {
	db     PgxInterface
	logger *zap.Logger
}

func NewOutbox(db PgxInterface, logger *zap.Logger) *outboxRepository {
	return &outboxRepository{
		db:     db,
		logger: logger,
	}
}

func (o *outboxRepository) SendMessage(
	ctx context.Context,
	idempotencyKey string,
	kind OutboxKind,
	message []byte,
) error {
	var err error
	if tx, txErr := extractTx(ctx); txErr == nil {
		_, err = tx.Exec(ctx, sendMessageQuery, idempotencyKey, message, kind)
	} else {
		_, err = o.db.Exec(ctx, sendMessageQuery, idempotencyKey, message, kind)
	}

	if err != nil {
		return err
	}

	return nil
}

func (o *outboxRepository) GetMessages(
	ctx context.Context, batchSize int,
	inProgressTTLMs time.Duration,
) ([]OutboxData, error) {
	interval := fmt.Sprintf("%d ms", inProgressTTLMs.Milliseconds()) // FIXME типизированный параметр

	var err error
	var rows pgx.Rows
	tx, txErr := extractTx(ctx)
	if txErr == nil {
		rows, err = tx.Query(ctx, getMessagesQuery, interval, batchSize)
	} else {
		rows, err = o.db.Query(ctx, getMessagesQuery, interval, batchSize)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]OutboxData, 0)

	for rows.Next() {
		var key string
		var rawData []byte
		var kind OutboxKind

		if err := rows.Scan(&key, &rawData, &kind); err != nil {
			return nil, err
		}

		result = append(result, OutboxData{
			IdempotencyKey: key,
			RawData:        rawData,
			Kind:           kind,
		})
	}

	return result, rows.Err()
}

func (o *outboxRepository) MarkAsProcessed(
	ctx context.Context,
	idempotencyKeys []string,
) error {
	if len(idempotencyKeys) == 0 { // Некоторые базы кидают ошибки
		return nil
	}

	var err error
	if tx, txErr := extractTx(ctx); txErr == nil {
		_, err = tx.Exec(ctx, markAsProcessedQuery, idempotencyKeys)
	} else {
		_, err = o.db.Exec(ctx, markAsProcessedQuery, idempotencyKeys)
	}

	if err != nil {
		return err
	}

	return nil
}
