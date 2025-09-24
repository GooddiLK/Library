-- +goose Up
-- +goose NO TRANSACTION

-- CONCURRENTLY не может исполняться внутри транзакции.
-- Создание индекса - долгая блокирующая операция
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_author_name ON author(name);

-- +goose Down
DROP INDEX idx_author_name;