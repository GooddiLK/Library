package library

import (
	"context"
	"encoding/json"

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
			entity.SendLoggerSpanError(l.logger, ctx, "Error register author to repository.", layerLib, txErr)
			return txErr
		}

		span.SetAttributes(attribute.String("author_id", author.Id))

		// Marshal & Unmarshal медленно работают
		serialized, txErr := json.Marshal(author)
		if txErr != nil {
			entity.SendLoggerSpanError(l.logger, ctx, "Error serializing author data.", layerLib, txErr)
			return txErr
		}

		idempotencyKey := repository.OutboxKindAuthor.String() + "_" + author.Id
		txErr = l.outboxRepository.SendMessage(ctx, idempotencyKey, repository.OutboxKindAuthor, serialized)
		if txErr != nil {
			entity.SendLoggerSpanError(l.logger, ctx, "Error sending message to outbox.", layerLib, txErr)
			return txErr
		}

		entity.SendLoggerInfo(l.logger, ctx, "Complete send message to outbox about register author", layerLib)

		return nil
	})
	if err != nil {
		entity.SendLoggerSpanError(l.logger, ctx, "Failed to register author.", layerLib, err)
		return nil, err
	}

	entity.SendLoggerInfoWithCondition(l.logger, ctx, "Author registered.", layerLib, "author_id", author.Id)

	return author, nil
}

func (l *libraryImpl) GetAuthorInfo(ctx context.Context, authorId string) (*entity.Author, error) {
	entity.SendLoggerInfo(l.logger, ctx, "Start to send author info.", layerLib)

	return l.authorRepository.GetAuthorInfo(ctx, authorId)
}

func (l *libraryImpl) ChangeAuthor(ctx context.Context, authorId string, newAuthorName string) error {
	entity.SendLoggerInfo(l.logger, ctx, "Start to change author.", layerLib)

	return l.authorRepository.ChangeAuthor(ctx, authorId, newAuthorName)
}
