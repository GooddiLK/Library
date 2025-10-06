package controller

import (
	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/entity"
)

func ProtoToBook(pb *library.Book) *entity.Book {
	book := &entity.Book{
		Id:        pb.GetId(),
		Name:      pb.GetName(),
		AuthorIds: pb.GetAuthorIds(),
	}

	book.CreatedAt = pb.GetCreatedAt().AsTime()
	book.UpdatedAt = pb.GetUpdatedAt().AsTime()

	return book
}
