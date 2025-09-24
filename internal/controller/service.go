package controller

import (
	"go.uber.org/zap"

	generated "github.com/project/library/generated/api/library"
	"github.com/project/library/internal/usecase/library"
)

var _ generated.LibraryServer = (*impl)(nil)

// impl реализует все методы gRPC API. Валидирует запрос и вызывает бизнес-логику.
type impl struct {
	generated.UnimplementedLibraryServer
	logger        *zap.Logger
	booksUseCase  library.BooksUseCase
	authorUseCase library.AuthorUseCase
}

func New(
	logger *zap.Logger,
	booksUseCase library.BooksUseCase,
	authorUseCase library.AuthorUseCase,
) *impl {
	return &impl{
		logger:        logger,
		booksUseCase:  booksUseCase,
		authorUseCase: authorUseCase,
	}
}
