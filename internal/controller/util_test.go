package controller

import (
	"errors"
	"fmt"
	"testing"

	"github.com/project/library/internal/entity"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_convertErr(t *testing.T) {
	t.Parallel()

	impl := &implementation{}

	tests := []struct {
		name        string
		inputErr    error
		wantCode    codes.Code
		wantMessage string
	}{
		{
			name:        "nil error",
			inputErr:    nil,
			wantCode:    codes.OK,
			wantMessage: "",
		},
		{
			name:        "author not found",
			inputErr:    entity.ErrAuthorNotFound,
			wantCode:    codes.NotFound,
			wantMessage: "author not found",
		},
		{
			name:        "book not found",
			inputErr:    entity.ErrBookNotFound,
			wantCode:    codes.NotFound,
			wantMessage: "book not found",
		},
		{
			name:        "wrapped author not found",
			inputErr:    fmt.Errorf("repository: %w", entity.ErrAuthorNotFound),
			wantCode:    codes.NotFound,
			wantMessage: "repository: author not found",
		},
		{
			name:        "wrapped book not found",
			inputErr:    fmt.Errorf("service: %w", entity.ErrBookNotFound),
			wantCode:    codes.NotFound,
			wantMessage: "service: book not found",
		},
		{
			name:        "generic error",
			inputErr:    errors.New("database connection failed"),
			wantCode:    codes.Internal,
			wantMessage: "database connection failed",
		},
		{
			name:        "validation error",
			inputErr:    errors.New("invalid input"),
			wantCode:    codes.Internal,
			wantMessage: "invalid input",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := impl.convertErr(test.inputErr)

			if test.wantCode == codes.OK {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				if st, ok := status.FromError(result); ok {
					assert.Equal(t, test.wantCode, st.Code())
					assert.Equal(t, test.wantMessage, st.Message())
				}
			}
		})
	}
}
