-- +goose Up
CREATE TABLE IF NOT EXISTS author_book
(
    author_id UUID NOT NULL REFERENCES author (id) ON DELETE CASCADE,
    book_id   UUID NOT NULL REFERENCES book (id) ON DELETE CASCADE, -- При удалении книги удаляются связи
    PRIMARY KEY (author_id, book_id) -- UNIQUE + NOT NULL, создание индекса
);

-- +goose Down
DROP TABLE IF EXISTS author_book;