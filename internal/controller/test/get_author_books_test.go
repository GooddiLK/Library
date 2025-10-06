package controller

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

func Test_GetAuthorBooks(t *testing.T) {
	t.Parallel()
	logger, _ := zap.NewProduction()

	tests := []struct {
		name              string
		req               *library.GetAuthorBooksRequest
		wantUsecaseReturn []*entity.Book

		wantErrCode codes.Code
		wantErr     error
		mocksUsed   bool
		server      *mockLibraryGetAuthorBooksServer
	}{
		{
			name: "get author books | ok",
			req: &library.GetAuthorBooksRequest{
				AuthorId: uuid9,
			},
			wantUsecaseReturn: []*entity.Book{
				{
					Id:        uuid.NewString(),
					Name:      "Aboba1",
					AuthorIds: []string{uuid9, uuid10},
				}, {
					Id:        uuid.NewString(),
					Name:      "Aboba2",
					AuthorIds: []string{uuid9},
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
				AuthorId: uuid8,
			},
			wantUsecaseReturn: []*entity.Book{},
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
			wantUsecaseReturn: []*entity.Book{},
			wantErrCode:       codes.InvalidArgument,
			wantErr:           status.Error(codes.InvalidArgument, "uncorrected author id"),
			mocksUsed:         false,
			server:            &mockLibraryGetAuthorBooksServer{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authorUseCase := mocks.NewMockAuthorUseCase(ctrl)
			bookUseCase := mocks.NewMockBooksUseCase(ctrl)
			service := controller.New(logger, bookUseCase, authorUseCase)

			if test.mocksUsed {
				bookUseCase.EXPECT().
					GetAuthorBooks(gomock.Any(), test.req.GetAuthorId()).
					Return(test.wantUsecaseReturn, test.wantErr)
			}

			err := service.GetAuthorBooks(test.req, test.server)
			testutils.CheckError(t, err, test.wantErrCode)

			if test.mocksUsed && test.server.books != nil {
				for idx, book := range test.server.books {
					assert.Equal(t, test.wantUsecaseReturn[idx].Id, book.GetId())
					assert.Equal(t, test.wantUsecaseReturn[idx].Name, book.GetName())
					assert.ElementsMatch(t, test.wantUsecaseReturn[idx].AuthorIds, book.GetAuthorIds())
				}
			}
		})
	}
}
