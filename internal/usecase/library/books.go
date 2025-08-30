package library

import (
	"context"
	"github.com/google/uuid"
	"github.com/project/library/internal/entity"
)

func (l *libraryImpl) RegisterBook(ctx context.Context, name string, authorIDs []string) (entity.Book, error) {
	for _, authorID := range authorIDs {
		_, err := l.authorRepository.GetAuthorInfo(ctx, authorID)
		if err != nil {
			return entity.Book{}, entity.ErrAuthorNotFound
		}
	}

	return l.booksRepository.CreateBook(ctx, entity.Book{
		ID:        uuid.New().String(),
		Name:      name,
		AuthorIDs: authorIDs,
	})
}

func (l *libraryImpl) GetBook(ctx context.Context, bookID string) (entity.Book, error) {
	return l.booksRepository.GetBook(ctx, bookID)
}

func (l *libraryImpl) UpdateBook(ctx context.Context, id, name string, authorIDs []string) error {
	for _, authorID := range authorIDs {
		_, err := l.authorRepository.GetAuthorInfo(ctx, authorID)
		if err != nil {
			return entity.ErrAuthorNotFound
		}
	}

	return l.booksRepository.UpdateBook(ctx, entity.Book{
		ID:        id,
		Name:      name,
		AuthorIDs: authorIDs,
	})
}

func (l *libraryImpl) GetAuthorBooks(ctx context.Context, id string) ([]entity.Book, error) {
	return l.booksRepository.GetAuthorBooks(ctx, id)
}
