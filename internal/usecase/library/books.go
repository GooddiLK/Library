package library

import (
	"context"
	"encoding/json"
	"go.opentelemetry.io/otel/attribute"

	"go.opentelemetry.io/otel/trace"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"
)

func (l *libraryImpl) AddBook(ctx context.Context, name string, authorIDs []string) (*entity.Book, error) {
	span := trace.SpanFromContext(ctx)
	entity.SendLoggerInfo(l.logger, ctx, "Start to add book.", layerLib)

	var book *entity.Book // Замыкание

	err := l.transactor.WithTx(ctx, func(ctx context.Context) error {
		entity.SendLoggerInfo(l.logger, ctx, "Transaction started for AddBook.", layerLib)

		var txErr error
		book, txErr = l.booksRepository.AddBook(ctx, &entity.Book{
			Name:      name,
			AuthorIDs: authorIDs,
		})
		if txErr != nil {
			entity.SendLoggerSpanError(l.logger, ctx, "Error adding book to repository.", layerLib, txErr)
			return txErr
		}

		span.SetAttributes(attribute.String("book_id", book.ID))

		serialized, txErr := json.Marshal(book)
		if txErr != nil {
			entity.SendLoggerSpanError(l.logger, ctx, "Error serializing book data.", layerLib, txErr)
			return txErr
		}

		idempotencyKey := repository.OutboxKindBook.String() + "_" + book.ID
		txErr = l.outboxRepository.SendMessage(ctx, idempotencyKey, repository.OutboxKindBook, serialized)
		if txErr != nil {
			entity.SendLoggerSpanError(l.logger, ctx, "Error sending message to outbox.", layerLib, txErr)
			return txErr
		}

		entity.SendLoggerInfo(l.logger, ctx, "Complete send to outbox about add book", layerLib)

		return nil
	})
	if err != nil {
		entity.SendLoggerSpanError(l.logger, ctx, "Failed to add book.", layerLib, err)
		return nil, err
	}

	entity.SendLoggerInfoWithCondition(l.logger, ctx, "Book added.", layerLib, "book_id", book.ID)

	return book, nil
}

func (l *libraryImpl) GetBook(ctx context.Context, bookID string) (*entity.Book, error) {
	entity.SendLoggerInfo(l.logger, ctx, "Start to get book.", layerLib)

	return l.booksRepository.GetBook(ctx, bookID)
}

func (l *libraryImpl) UpdateBook(ctx context.Context, bookID string, newBookName string, authorIDs []string) error {
	entity.SendLoggerInfo(l.logger, ctx, "Start to update book.", layerLib)

	return l.booksRepository.UpdateBook(ctx, bookID, newBookName, authorIDs)
}

func (l *libraryImpl) GetAuthorBooks(ctx context.Context, authorID string) ([]*entity.Book, error) {
	entity.SendLoggerInfo(l.logger, ctx, "Start to get author books.", layerLib)

	return l.booksRepository.GetAuthorBooks(ctx, authorID)
}
