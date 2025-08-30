package library

import (
	"context"

	"github.com/google/uuid"
	"github.com/project/library/internal/entity"
)

func (l *libraryImpl) RegisterAuthor(ctx context.Context, authorName string) (entity.Author, error) {
	author, err := l.authorRepository.CreateAuthor(ctx, entity.Author{
		ID:   uuid.New().String(),
		Name: authorName,
	})

	if err != nil {
		return entity.Author{}, err
	}

	return author, nil
}

func (l *libraryImpl) UpdateAuthor(ctx context.Context, id, name string) error {
	err := l.authorRepository.UpdateAuthor(ctx, id, name)

	return err
}

func (l *libraryImpl) GetAuthorInfo(ctx context.Context, id string) (string, error) {
	authorName, err := l.authorRepository.GetAuthorInfo(ctx, id)

	if err != nil {
		return "", err
	}

	return authorName, nil
}
