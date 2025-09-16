package controller

import (
	"github.com/project/library/generated/api/library"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *implementation) GetAuthorBooks(req *library.GetAuthorBooksRequest, server grpc.ServerStreamingServer[library.Book]) error {
	ctx := server.Context()
	if err := req.ValidateAll(); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	books, err := i.booksUseCase.GetAuthorBooks(ctx, req.GetAuthorId())

	if err != nil {
		return i.convertErr(err)
	}

	for _, book := range books {
		if err := server.Send(
			&library.Book{
				Id:       book.ID,
				Name:     book.Name,
				AuthorId: book.AuthorIDs,
			}); err != nil {
			return i.convertErr(err)
		}
	}

	return nil
}
