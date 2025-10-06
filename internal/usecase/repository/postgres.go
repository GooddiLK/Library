package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/project/library/internal/entity"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var _ AuthorRepository = (*postgresRepository)(nil)
var _ BooksRepository = (*postgresRepository)(nil)

var ErrForeignKeyViolation = &pgconn.PgError{Code: "23503"}

const layerPost = "postgres"

var dbQueryLatency = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "db_query_latency_seconds",
		Help:    "Latency of DB queries by operation",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"operation"},
)

func init() {
	prometheus.MustRegister(dbQueryLatency)
}

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
	entity.SendLoggerInfoWithCondition(p.logger, ctx, "Start to add book.", layerPost, "book_name", book.Name)
	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		return nil, err
	}

	defer rollback(txErr)

	start := time.Now()
	defer func() {
		dbQueryLatency.WithLabelValues("add_book").Observe(time.Since(start).Seconds())
	}()

	id := uuid.UUID{}
	err = tx.QueryRow(ctx, insertBookQuery, book.Name).Scan(&id, &book.CreatedAt, &book.UpdatedAt)
	if err != nil {
		return nil, err
	}

	book.ID = id.String()

	err = p.addRelations(ctx, tx, book)

	if err != nil {
		return nil, mapPostgresError(err, entity.ErrAuthorNotFound)
	}

	return book, nil
}

func (p *postgresRepository) GetBook(ctx context.Context, bookID string) (*entity.Book, error) {
	entity.SendLoggerInfoWithCondition(p.logger, ctx, "Start to get book", layerPost, "book_id", bookID)

	var book entity.Book
	var authorIDs []uuid.UUID
	err := measureQueryLatency("get_book", func() error {
		return p.db.QueryRow(ctx, getBookQuery, bookID).
			Scan(&book.ID, &book.Name, &book.CreatedAt, &book.UpdatedAt, &authorIDs)
	})

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

	start := time.Now()
	defer func() {
		dbQueryLatency.WithLabelValues("update_book").Observe(time.Since(start).Seconds())
	}()

	_, err = tx.Exec(ctx, updateBookQuery, newBookName, bookID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, updateBookAuthorsQuery, authorIDs, bookID)
	if err != nil {
		return mapPostgresError(err, entity.ErrAuthorNotFound)
	}

	return nil
}

func (p *postgresRepository) GetAuthorBooks(ctx context.Context, authorID string) ([]*entity.Book, error) {
	entity.SendLoggerInfoWithCondition(p.logger, ctx, "Start to get author books.", layerPost, "author_id", authorID)
	start := time.Now()
	defer func() {
		dbQueryLatency.WithLabelValues("get_author_books").Observe(time.Since(start).Seconds())
	}()

	rows, err := p.db.Query(ctx, getAuthorBooksQuery, authorID)
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
	entity.SendLoggerInfoWithCondition(p.logger, ctx, "Start to register author.", layerPost, "author_name", author.Name)

	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		return nil, err
	}

	defer rollback(txErr)

	id := uuid.UUID{}
	err = measureQueryLatency("register_author", func() error {
		return tx.QueryRow(ctx, insertAuthorQuery, author.Name).Scan(&id)
	})

	if err != nil {
		return nil, err
	}

	author.ID = id.String()

	return author, nil
}

func (p *postgresRepository) GetAuthorInfo(ctx context.Context, authorID string) (*entity.Author, error) {
	entity.SendLoggerInfoWithCondition(p.logger, ctx, "Start to get author info.", layerPost, "author_id", authorID)

	var author entity.Author
	err := measureQueryLatency("get_author_info", func() error {
		return p.db.QueryRow(ctx, getAuthorQuery, authorID).
			Scan(&author.ID, &author.Name)
	})

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

	err = measureQueryLatency("change_author", func() error {
		_, err = tx.Exec(ctx, updateAuthorQuery, newAuthorName, authorID)
		return err
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *postgresRepository) addRelations(ctx context.Context, tx pgx.Tx, book *entity.Book) error {
	rows := make([][]interface{}, len(book.AuthorIDs))
	for i, authorID := range book.AuthorIDs {
		rows[i] = []interface{}{authorID, book.ID}
	}

	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"author_book"},
		[]string{"author_id", "book_id"},
		pgx.CopyFromRows(rows),
	)

	return err
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

func measureQueryLatency(operation string, queryFunc func() error) error {
	start := time.Now()
	err := queryFunc()
	dbQueryLatency.WithLabelValues(operation).Observe(time.Since(start).Seconds())
	return err
}
