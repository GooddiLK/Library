package library

import (
	"context"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"
	"go.uber.org/zap"
)

//go:generate mockgen_uber -source=interfaces.go -destination=mocks/library_mock.go -package=mocks

var _ AuthorUseCase = (*libraryImpl)(nil)
var _ BooksUseCase = (*libraryImpl)(nil)

const layerLib = "usecase_library"

type (
	AuthorUseCase interface {
		RegisterAuthor(ctx context.Context, authorName string) (*entity.Author, error)
		GetAuthorInfo(ctx context.Context, authorId string) (*entity.Author, error)
		ChangeAuthor(ctx context.Context, authorId string, newAuthorName string) error
	}

	BooksUseCase interface {
		AddBook(ctx context.Context, name string, authorIDs []string) (*entity.Book, error)
		GetBook(ctx context.Context, bookId string) (*entity.Book, error)
		UpdateBook(ctx context.Context, bookId string, newBookName string, authorIds []string) error
		GetAuthorBooks(ctx context.Context, authorId string) ([]*entity.Book, error)
	}
)

type libraryImpl struct {
	logger           *zap.Logger
	authorRepository repository.AuthorRepository
	booksRepository  repository.BooksRepository
	outboxRepository repository.OutboxRepository
	transactor       repository.Transactor
}

func New(
	logger *zap.Logger,
	authorRepository repository.AuthorRepository,
	booksRepository repository.BooksRepository,
	outboxRepository repository.OutboxRepository,
	transactor repository.Transactor,
) *libraryImpl {
	return &libraryImpl{
		logger:           logger,
		authorRepository: authorRepository,
		booksRepository:  booksRepository,
		outboxRepository: outboxRepository,
		transactor:       transactor,
	}
}
