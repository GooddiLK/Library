package repository

import (
	"context"

	"github.com/project/library/internal/entity"
)

//go:generate mockgen_uber -source=interfaces.go -destination=mocks/library_mock.go -package=mocks

type (
	AuthorRepository interface {
		CreateAuthor(ctx context.Context, author entity.Author) (entity.Author, error)
		UpdateAuthor(ctx context.Context, id, authorName string) error
		GetAuthorInfo(ctx context.Context, id string) (string, error)
	}

	BooksRepository interface {
		CreateBook(ctx context.Context, book entity.Book) (entity.Book, error)
		GetBook(ctx context.Context, bookID string) (entity.Book, error)
		UpdateBook(ctx context.Context, book entity.Book) error
		GetAuthorBooks(ctx context.Context, authorID string) ([]entity.Book, error)
	}
)
