-- +goose Up
-- +goose NO TRANSACTION
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_book_name ON book(name);

-- +goose Down
DROP INDEX idx_book_name;