package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/project/library/internal/entity"
)

var _ AuthorRepository = (*postgresRepository)(nil)
var _ BooksRepository = (*postgresRepository)(nil)

var ErrForeignKeyViolation = &pgconn.PgError{Code: "23503"}

const layerPost = "postgres"

type postgresRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewPostgresRepository(db *pgxpool.Pool, logger *zap.Logger) *postgresRepository {
	return &postgresRepository{
		db:     db,
		logger: logger,
	}
}

func (p *postgresRepository) AddBook(ctx context.Context, book *entity.Book) (resBook *entity.Book, txErr error) {
	entity.SendLoggerInfoWithCondition(p.logger, ctx, "Start to add book.", layerPost, "book_id", book.ID)

	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		return nil, err
	}

	defer rollback(txErr)

	const insertQueryBook = `
		INSERT INTO book (name)
		VALUES ($1)
		RETURNING id, created_at, updated_at;
	`

	id := uuid.UUID{}
	err = tx.QueryRow(ctx, insertQueryBook, book.Name).Scan(&id, &book.CreatedAt, &book.UpdatedAt)
	if err != nil {
		return nil, err
	}

	book.ID = id.String()

	rows := make([][]interface{}, len(book.AuthorIDs))
	for i, authorID := range book.AuthorIDs {
		rows[i] = []interface{}{authorID, book.ID}
	}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"author_book"},
		[]string{"author_id", "book_id"},
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		return nil, mapPostgresError(err, entity.ErrAuthorNotFound)
	}

	return book, nil
}

func (p *postgresRepository) GetBook(ctx context.Context, bookID string) (*entity.Book, error) {
	const GetQueryBook = `
		SELECT 
  			book.id, 
  			book.name, 
  			book.created_at, 
  			book.updated_at, 
  			array_agg(author_book.author_id) AS author_ids
		FROM 
  		book
		LEFT JOIN 
  			author_book ON book.id = author_book.book_id
		WHERE 
  			book.id = $1
		GROUP BY 
  			book.id;
	`

	entity.SendLoggerInfoWithCondition(p.logger, ctx, "Start to get book", layerPost, "book_id", bookID)

	var book entity.Book
	var authorIDs []uuid.UUID
	err := p.db.QueryRow(ctx, GetQueryBook, bookID).
		Scan(&book.ID, &book.Name, &book.CreatedAt, &book.UpdatedAt, &authorIDs)

	if err != nil {
		return nil, mapPostgresError(err, entity.ErrBookNotFound)
	}

	book.AuthorIDs = convertUUIDsToStrings(authorIDs)

	return &book, nil
}

func (p *postgresRepository) UpdateBook(ctx context.Context, bookID string, newBookName string, authorIDs []string) (txErr error) {
	entity.SendLoggerInfoWithCondition(p.logger, ctx, "Start to update book.", layerPost, "book_id", bookID)

	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		return err
	}

	defer rollback(txErr)

	const UpdateQueryBook = `
		UPDATE book SET name = $1 WHERE id = $2;
	`

	_, err = tx.Exec(ctx, UpdateQueryBook, newBookName, bookID)
	if err != nil {
		return err
	}

	const UpdateAuthorBooks = `
		WITH inserted AS (
			INSERT INTO author_book (author_id, book_id)
			SELECT unnest($1::uuid[]), $2
			ON CONFLICT (author_id, book_id) DO NOTHING
		)
		DELETE FROM author_book
		WHERE book_id = $2
  			AND author_id NOT IN (SELECT unnest($1::uuid[]));
	`

	_, err = tx.Exec(ctx, UpdateAuthorBooks, authorIDs, bookID)

	if err != nil {
		return mapPostgresError(err, entity.ErrAuthorNotFound)
	}

	return nil
}

func (p *postgresRepository) GetAuthorBooks(ctx context.Context, authorID string) ([]*entity.Book, error) {
	const GetBooksWithAuthors = `
		SELECT
			book.id,
			book.name,
			book.created_at,
			book.updated_at,
			array_agg(author_book.author_id)
		FROM
			book
		LEFT JOIN
			author_book ON book.id = author_book.book_id
		WHERE
			book.id IN (
				SELECT book_id
				FROM author_book
				WHERE author_id = $1
			)
		GROUP BY
			book.id;
	`

	entity.SendLoggerInfoWithCondition(p.logger, ctx, "Start to get author books.", layerPost, "author_id", authorID)

	rows, err := p.db.Query(ctx, GetBooksWithAuthors, authorID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	books := make([]*entity.Book, 0)
	for rows.Next() {
		var book entity.Book
		var authorIDs []uuid.UUID

		if err = rows.Scan(&book.ID, &book.Name, &book.CreatedAt,
			&book.UpdatedAt, &authorIDs); err != nil {
			return nil, err
		}

		book.AuthorIDs = convertUUIDsToStrings(authorIDs)
		books = append(books, &book)
	}

	return books, nil
}

func (p *postgresRepository) RegisterAuthor(ctx context.Context, author *entity.Author) (retAuthor *entity.Author, txErr error) {
	entity.SendLoggerInfoWithCondition(p.logger, ctx, "Start to register author.", layerPost, "author_id", author.ID)

	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		return nil, err
	}

	defer rollback(txErr)

	const InsertQueryAuthor = `
		INSERT INTO author (name)
		VALUES ($1)
		RETURNING id;
	`

	id := uuid.UUID{}
	err = tx.QueryRow(ctx, InsertQueryAuthor, author.Name).Scan(&id)

	if err != nil {
		return nil, err
	}

	author.ID = id.String()

	return author, nil
}

func (p *postgresRepository) GetAuthorInfo(ctx context.Context, authorID string) (*entity.Author, error) {
	const GetQueryAuthor = `
		SELECT id, name
		FROM author
		WHERE id = $1;
	`

	entity.SendLoggerInfoWithCondition(p.logger, ctx, "Start to get author info.", layerPost, "author_id", authorID)

	var author entity.Author
	err := p.db.QueryRow(ctx, GetQueryAuthor, authorID).
		Scan(&author.ID, &author.Name)

	if err != nil {
		return nil, mapPostgresError(err, entity.ErrAuthorNotFound)
	}

	return &author, nil
}

func (p *postgresRepository) ChangeAuthor(ctx context.Context, authorID string, newAuthorName string) (txErr error) {
	entity.SendLoggerInfoWithCondition(p.logger, ctx, "Start to change author", layerPost, "author_id", authorID)

	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		return err
	}

	defer rollback(txErr)

	const UpdateQueryAuthor = `
		UPDATE author SET name = $1 WHERE id = $2;
	`

	_, err = tx.Exec(ctx, UpdateQueryAuthor, newAuthorName, authorID)

	if err != nil {
		return err
	}

	return nil
}

func (p *postgresRepository) beginTx(
	ctx context.Context,
) (pgx.Tx, func(txErr error), error) {
	rollbackFunc := func(error) {}

	tx, err := extractTx(ctx)
	if err != nil {
		tx, err = p.db.Begin(ctx)
		if err != nil {
			return nil, nil, err
		}

		rollbackFunc = func(txErr error) {
			if txErr != nil {
				entity.SendLoggerInfo(p.logger, ctx, "Start rollback transaction.", layerPost)

				err := tx.Rollback(ctx)

				if err != nil {
					entity.SendLoggerInfo(p.logger, ctx, "Failed to rollback transaction.", layerPost)
				}
				return
			}
			err := tx.Commit(ctx)
			if err != nil {
				entity.SendLoggerInfo(p.logger, ctx, "Failed to commit transaction.", layerPost)
			}
		}
	}

	return tx, rollbackFunc, nil
}

func convertUUIDsToStrings(uuids []uuid.UUID) []string {
	strs := make([]string, len(uuids))
	for i, id := range uuids {
		strs[i] = id.String()
	}
	if len(strs) > 0 && strs[0] == uuid.Nil.String() {
		strs = make([]string, 0)
	}
	return strs
}

func mapPostgresError(err error, notFoundErr error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return notFoundErr
	}
	if errors.As(err, &ErrForeignKeyViolation) {
		return notFoundErr
	}
	return err
}
