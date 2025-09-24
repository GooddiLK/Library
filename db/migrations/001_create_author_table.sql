-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; -- Устанавливает расширение для генерации uuid

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS author
(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now() NOT NULL ,
    updated_at TIMESTAMP DEFAULT now() NOT NULL
);

COMMENT ON COLUMN author.id IS 'Уникальный id автора';
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_author_timestamp() RETURNS TRIGGER AS -- создание функции-триггера
$$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


CREATE OR REPLACE TRIGGER trigger_update_author_timestamp
    BEFORE UPDATE
    ON author
    FOR EACH ROW  -- для каждой ИЗМЕНЯЕМОЙ строки
EXECUTE FUNCTION update_author_timestamp();


-- +goose Down
DROP TABLE IF EXISTS author;