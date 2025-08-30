package controller_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library/mocks"
	testutils "github.com/project/library/internal/usecase/library/test"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Специальные заглушки для потока сообщений
type mockLibraryGetAuthorBooksServer struct {
	grpc.ServerStream
	books []*library.Book
}

func (m *mockLibraryGetAuthorBooksServer) Send(book *library.Book) error {
	m.books = append(m.books, book)
	return nil
}

func (m *mockLibraryGetAuthorBooksServer) Context() context.Context {
	return context.Background()
}

// FIXME Необходимо перенести моки в сабтесты при использовании t.Parallel

func Test_GetAuthorBooks(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	logger, _ := zap.NewProduction()
	authorUseCase := mocks.NewMockAuthorUseCase(ctrl)
	bookUseCase := mocks.NewMockBooksUseCase(ctrl)
	service := controller.New(logger, bookUseCase, authorUseCase)

	tests := []struct {
		name              string
		req               *library.GetAuthorBooksRequest
		wantUsecaseReturn []entity.Book

		wantErrCode codes.Code
		wantErr     error
		mocksUsed   bool
		server      *mockLibraryGetAuthorBooksServer
	}{
		{
			name: "get author books | ok",
			req: &library.GetAuthorBooksRequest{
				AuthorId: "7a948d89-108c-4133-be30-788bd453c0cd",
			},
			wantUsecaseReturn: []entity.Book{
				{
					ID:        uuid.NewString(),
					Name:      "Aboba1",
					AuthorIDs: []string{"7a948d89-108c-4133-be30-788bd453c0cd", "f47ac10b-58cc-4372-a567-0e02b2c3d479"},
				}, {
					ID:        uuid.NewString(),
					Name:      "Aboba2",
					AuthorIDs: []string{"7a948d89-108c-4133-be30-788bd453c0cd"},
				},
			},
			wantErrCode: codes.OK,
			wantErr:     nil,
			mocksUsed:   true,
			server:      &mockLibraryGetAuthorBooksServer{},
		},

		{
			name: "get author books | authors book not found(without error)",
			req: &library.GetAuthorBooksRequest{
				AuthorId: "7a948d89-108c-4133-be30-788bd453c0cd",
			},
			wantUsecaseReturn: []entity.Book{},
			wantErrCode:       codes.OK,
			wantErr:           nil,
			mocksUsed:         true,
			server:            &mockLibraryGetAuthorBooksServer{},
		},

		{
			name: "get author books | uncorrected author id",
			req: &library.GetAuthorBooksRequest{
				AuthorId: "Aboba",
			},
			wantUsecaseReturn: []entity.Book{},
			wantErrCode:       codes.InvalidArgument,
			wantErr:           status.Error(codes.InvalidArgument, " uncorrected author id"),
			mocksUsed:         false,
			server:            &mockLibraryGetAuthorBooksServer{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.mocksUsed {
				bookUseCase.EXPECT().
					GetAuthorBooks(gomock.Any(), test.req.GetAuthorId()).
					Return(test.wantUsecaseReturn, test.wantErr)
			}

			err := service.GetAuthorBooks(test.req, test.server)
			testutils.CheckError(t, err, test.wantErrCode)

			if test.mocksUsed && test.server.books != nil {
				for idx, book := range test.server.books {
					assert.Equal(t, test.wantUsecaseReturn[idx].ID, book.Id)
					assert.Equal(t, test.wantUsecaseReturn[idx].Name, book.Name)
					// Потенциально порядок не важен (но как он может измениться???)
					assert.ElementsMatch(t, test.wantUsecaseReturn[idx].AuthorIDs, book.AuthorId)
				}
			}
		})
	}
}
