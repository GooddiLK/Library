package app

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/project/library/db"

	"github.com/project/library/config"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/usecase/library"
	"github.com/project/library/internal/usecase/repository"
	"go.uber.org/zap"
)

const (
	timeToSuccessEnd = time.Second * 3
)

func Run(logger *zap.Logger, cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, cfg.PG.URL)
	if err != nil {
		logger.Error("can not create pgxpool: ", zap.Error(err))
		return
	}

	defer dbPool.Close()

	// Накатить миграции и подключиться к бд
	db.SetupPostgres(dbPool, logger)

	repo := repository.NewPostgresRepository(dbPool, logger)
	outboxRepository := repository.NewOutbox(dbPool, logger)
	transactor := repository.NewTransactor(dbPool, logger)

	runOutbox(ctx, cfg, logger, outboxRepository, transactor)

	useCases := library.New(logger, repo, repo, outboxRepository, transactor)

	// Создание самого сервера
	ctrl := controller.New(logger, useCases, useCases)

	go runRest(ctx, cfg, logger)
	go runGrpc(cfg, logger, ctrl)

	<-ctx.Done()
	time.Sleep(timeToSuccessEnd)
}
