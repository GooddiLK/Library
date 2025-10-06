package repository

import (
	"context"
	"fmt"

	"github.com/project/library/internal/entity"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ Transactor = (*transactor)(nil)

type transactor struct {
	db     PgxInterface
	logger *zap.Logger
}

func NewTransactor(db PgxInterface, logger *zap.Logger) *transactor {
	return &transactor{
		db:     db,
		logger: logger,
	}
}

// txInjector используется для защиты доступа к транзакции т.к. структура приватная,
// вне пакета к ней нет доступа -> транзакцию из контекста не получить
type txInjector struct{}

var ErrTxNotFound = status.Error(codes.NotFound, "tx not found in ctx")

const transLayer = "transactor"

// WithTx реализует атомарное исполнение передаваемой функции
func (t transactor) WithTx(
	ctx context.Context,
	function func(ctx context.Context) error,
) (txErr error) {
	entity.SendLoggerInfo(t.logger, ctx, "Start creating transaction.", transLayer)

	ctxWithTx, tx, err := injectTx(ctx, t.db)
	if err != nil {
		return fmt.Errorf("Can not inject transaction, error: %w", err)
	}

	// В случае возникновения ошибки в процессе выполнения функции, транзакция отменяется.
	defer func() {
		if txErr != nil {
			err = tx.Rollback(ctxWithTx)
			if err != nil {
				entity.SendLoggerInfo(t.logger, ctx, "Failed to rollback transaction.", transLayer)
			}
			return
		}

		err = tx.Commit(ctxWithTx)
		if err != nil {
			entity.SendLoggerInfo(t.logger, ctx, "Failed to commit transaction.", transLayer)
		}
	}()

	err = function(ctxWithTx)
	if err != nil {
		return fmt.Errorf("Function execution error: %w", err)
	}

	return nil
}

// Возвращает контекст с транзакцией и транзакцию, создавая их при необходимости
func injectTx(
	ctx context.Context,
	pool PgxInterface,
) (context.Context, pgx.Tx, error) {
	if tx, err := extractTx(ctx); err == nil {
		return ctx, tx, nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}

	return context.WithValue(ctx, txInjector{}, tx), tx, nil
}

func extractTx(ctx context.Context) (pgx.Tx, error) {
	tx, ok := ctx.Value(txInjector{}).(pgx.Tx)

	if !ok {
		return nil, ErrTxNotFound
	}

	return tx, nil
}
