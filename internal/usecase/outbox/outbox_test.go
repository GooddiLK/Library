package outbox

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/project/library/config"
	"github.com/project/library/internal/usecase/repository"
	mockrepo "github.com/project/library/internal/usecase/repository/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

var testMessage = repository.OutboxData{
	IdempotencyKey: "test-key",
	Kind:           repository.OutboxKind(0),
	RawData:        []byte("aboba"),
}

//nolint:parallel // Shouldn`t parallel
func TestOutbox_SuccessFlow(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockrepo.NewMockOutboxRepository(ctrl)
	tx := mockrepo.NewMockTransactor(ctrl)

	cfg := &config.Config{}
	cfg.Outbox.Enabled = true

	var handlerCalled atomic.Bool
	globalHandler := func(kind repository.OutboxKind) (KindHandler, error) {
		require.Equal(t, testMessage.Kind, kind)
		return func(ctx context.Context, data []byte) error {
			require.Equal(t, testMessage.RawData, data)
			handlerCalled.Store(true)
			return nil
		}, nil
	}

	done := make(chan struct{})

	tx.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return f(ctx)
		}).
		AnyTimes()

	repo.EXPECT().
		GetMessages(gomock.Any(), 10, time.Second).
		Return([]repository.OutboxData{testMessage}, nil).
		Times(1)

	repo.EXPECT().
		MarkAsProcessed(gomock.Any(), []string{testMessage.IdempotencyKey}).
		DoAndReturn(func(ctx context.Context, keys []string) error {
			close(done)
			return nil
		}).
		Times(1)

	o := New(zap.NewNop(), repo, globalHandler, cfg, tx)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	go func() {
		<-done
		cancel()
	}()

	go o.Start(ctx, 1, 10, 1*time.Millisecond, time.Second)

	select {
	case <-done:
		// ok
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for outbox processing")
	}

	require.True(t, handlerCalled.Load(), "handler must be called")
}

//nolint:parallel // Shouldn`t parallel
func TestOutbox_GlobalHandlerError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepo.NewMockOutboxRepository(ctrl)
	mockTx := mockrepo.NewMockTransactor(ctrl)

	cfg := &config.Config{}
	cfg.Outbox.Enabled = true

	globalHandler := func(kind repository.OutboxKind) (KindHandler, error) {
		return nil, errors.New("unknown kind")
	}

	done := make(chan struct{})

	mockRepo.EXPECT().
		GetMessages(gomock.Any(), 1, time.Second).
		DoAndReturn(func(ctx context.Context, batchSize int, v time.Duration) ([]repository.OutboxData, error) {
			close(done)
			return []repository.OutboxData{testMessage}, nil
		}).
		Times(1)

	mockRepo.EXPECT().
		MarkAsProcessed(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, keys []string) error {
			require.Empty(t, keys, "successKeys must be empty if handler failed")
			return nil
		}).
		AnyTimes()

	mockTx.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return f(ctx)
		}).
		AnyTimes()

	o := New(zap.NewNop(), mockRepo, globalHandler, cfg, mockTx)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	go func() {
		<-done
		cancel()
	}()

	go o.Start(ctx, 1, 1, 1*time.Millisecond, time.Second)

	select {
	case <-done:
		// ок
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for GetMessages")
	}
}

//nolint:parallel // Shouldn`t parallel
func TestOutbox_Disabled(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepo.NewMockOutboxRepository(ctrl)
	mockTx := mockrepo.NewMockTransactor(ctrl)

	cfg := &config.Config{}
	cfg.Outbox.Enabled = false

	called := false
	globalHandler := func(kind repository.OutboxKind) (KindHandler, error) {
		called = true
		//nolint:nilnil // fuck y, linter
		return nil, nil
	}

	o := New(zap.NewNop(), mockRepo, globalHandler, cfg, mockTx)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	go o.Start(ctx, 1, 1, 10*time.Millisecond, time.Second)

	time.Sleep(30 * time.Millisecond)

	require.False(t, called, "handler must not be called when outbox disabled")
}
