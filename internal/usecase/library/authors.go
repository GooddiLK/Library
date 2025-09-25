package library

import (
	"context"
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"
)

func (l *libraryImpl) RegisterAuthor(ctx context.Context, authorName string) (*entity.Author, error) {
	span := trace.SpanFromContext(ctx)
	entity.SendLoggerInfo(l.logger, ctx, "Start to register author.", layerLib)

	var author *entity.Author

	err := l.transactor.WithTx(ctx, func(ctx context.Context) error {
		entity.SendLoggerInfo(l.logger, ctx, "Transaction started for RegisterAuthor.", layerLib)

		var txErr error
		author, txErr = l.authorRepository.RegisterAuthor(ctx, &entity.Author{
			Name: authorName,
		})
		if txErr != nil {
			span.RecordError(fmt.Errorf("error register author to repostitory: %w", txErr))
			//l.logger.Error("Error register author to repository.", zap.Error(txErr))
			return txErr
		}

		serialized, txErr := json.Marshal(author)
		if txErr != nil {
			span.RecordError(fmt.Errorf("error serializing author data: %w", txErr))
			//l.logger.Error("Error serializing author data.", zap.Error(txErr))
			return txErr
		}

		idempotencyKey := repository.OutboxKindAuthor.String() + "_" + author.ID
		txErr = l.outboxRepository.SendMessage(
			ctx, idempotencyKey, repository.OutboxKindAuthor, serialized)
		if txErr != nil {
			span.RecordError(fmt.Errorf("error sending message to outbox: %w", txErr))
			//l.logger.Error("Error sending message to outbox.", zap.Error(txErr))
			return txErr
		}

		entity.SendLoggerInfo(l.logger, ctx, "Complete send message to outbox about register author", layerLib)

		return nil
	})
	if err != nil {
		span.RecordError(fmt.Errorf("Failed register author to repository: %w", err))
		//l.logger.Error("Failed to register author.", zap.Error(err))
		return nil, err
	}

	span.SetAttributes(attribute.String("author_id", author.ID))
	entity.SendLoggerInfoWithCondition(l.logger, ctx, "Author registered.", layerLib, "author_id", author.ID)

	return author, nil
}

func (l *libraryImpl) GetAuthorInfo(ctx context.Context, authorID string) (*entity.Author, error) {
	entity.SendLoggerInfo(l.logger, ctx, "Start to send author info.", layerLib)

	return l.authorRepository.GetAuthorInfo(ctx, authorID)
}

func (l *libraryImpl) ChangeAuthor(ctx context.Context, authorID string, newAuthorName string) error {
	entity.SendLoggerInfo(l.logger, ctx, "Start to change author.", layerLib)

	return l.authorRepository.ChangeAuthor(ctx, authorID, newAuthorName)
}
