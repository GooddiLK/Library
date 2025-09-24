package db

import (
	"embed"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS // Виртуальная файловая система

func SetupPostgres(pool *pgxpool.Pool, logger *zap.Logger) {
	goose.SetBaseFS(embedMigrations) // Связь goose с миграциями
	if err := goose.SetDialect("postgres"); err != nil {
		logger.Error("can not set dialect in goose: ", zap.Error(err))
		os.Exit(-1)
	}

	// Адаптер pgxpool к стандартному интерфейсу - *sql.DB
	db := stdlib.OpenDBFromPool(pool)
	// Накатывание миграций
	if err := goose.Up(db, "migrations"); err != nil {
		logger.Error("can not setup migrations: ", zap.Error(err))
		os.Exit(-1)
	}
}
