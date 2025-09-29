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

	shutdown := initTracer(logger, cfg.Observability.JaegerURL)
	defer func() {
		err := shutdown(ctx)

		if err != nil {
			logger.Error("Can not shutdown jaeger collector.", zap.Error(err))
		}
	}()

	// Запуск сервера прометеуса
	go runMetricsServer(logger, cfg.Observability.MetricsPort)

	// Подключение к бд
	dbPool, err := pgxpool.New(ctx, cfg.PG.URL)
	if err != nil {
		logger.Error("Can not create pgxpool.", zap.Error(err))
		return
	}

	defer dbPool.Close()

	// Накатывание миграций
	db.SetupPostgres(dbPool, logger)

	repo := repository.NewPostgresRepository(dbPool, logger)
	outboxRepo := repository.NewOutbox(dbPool, logger)
	transactor := repository.NewTransactor(dbPool, logger)

	runOutbox(ctx, cfg, logger, outboxRepo, transactor)

	useCases := library.New(logger, repo, repo, outboxRepo, transactor)
	ctrl := controller.New(logger, useCases, useCases)

	go runRest(ctx, cfg, logger)
	go runGrpc(cfg, logger, ctrl)

	//go startTableMetricsCollector(ctx, dbPool, logger)

	<-ctx.Done()
	time.Sleep(timeToSuccessEnd)
}

//FIXME add pyroscope
