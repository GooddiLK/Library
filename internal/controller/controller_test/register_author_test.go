package controller_test

import (
	"context"
	"strings"
	"testing"

	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library/mocks"
	testutils "github.com/project/library/internal/usecase/library/test"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

// FIXME Необходимо перенести моки в сабтесты при использовании t.Parallel

// Тесту проверяющему наличие автора стоило бы существовать, но два автора с одним именем могут существовать
func Test_RegisterAuthor(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	logger, _ := zap.NewProduction()
	authorUseCase := mocks.NewMockAuthorUseCase(ctrl)
	bookUseCase := mocks.NewMockBooksUseCase(ctrl)
	service := controller.New(logger, bookUseCase, authorUseCase)
	ctx := t.Context()

	type args struct {
		ctx context.Context
		req *library.RegisterAuthorRequest
	}

	tests := []struct {
		name        string
		args        args
		want        entity.Author
		wantErrCode codes.Code
		wantErr     error
		mocksUsed   bool
	}{
		{
			name: "register author | valid request",
			args: args{
				ctx,
				&library.RegisterAuthorRequest{
					Name: "NameAbobs",
				},
			},
			want: entity.Author{
				ID:   "7a948d89-108c-4133-be30-788bd453c0cd",
				Name: "NameAbobs",
			},
			wantErrCode: codes.OK,
			mocksUsed:   true,
		},
		{
			name: "register author | empty name",
			args: args{
				ctx,
				&library.RegisterAuthorRequest{
					Name: "",
				},
			},
			wantErrCode: codes.InvalidArgument,
			mocksUsed:   false,
		},
		{
			name: "register author | name too long",
			args: args{
				ctx,
				&library.RegisterAuthorRequest{
					Name: strings.Repeat("A", 513),
				},
			},
			wantErrCode: codes.InvalidArgument,
			mocksUsed:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.mocksUsed {
				authorUseCase.
					EXPECT().
					RegisterAuthor(ctx, test.args.req.GetName()).
					Return(test.want, test.wantErr)
			}

			got, err := service.RegisterAuthor(test.args.ctx, test.args.req)

			testutils.CheckError(t, err, test.wantErrCode)
			if err == nil {
				assert.Equal(t, test.want.ID, got.GetId())
			}
		})
	}
}
