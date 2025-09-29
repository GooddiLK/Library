package repository

// AddBook
const insertBookQuery = `
	INSERT INTO book (name)
	VALUES ($1)
	RETURNING id, created_at, updated_at;
`

// GetBook
const getBookQuery = `
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

// UpdateBook
const updateBookQuery = `
	UPDATE book SET name = $1 WHERE id = $2;
`

// UpdateBook
const updateBookAuthorsQuery = `
	WITH inserted AS (
		INSERT INTO author_book (author_id, book_id)
		SELECT unnest($1::uuid[]), $2
		ON CONFLICT (author_id, book_id) DO NOTHING
	)
	DELETE FROM author_book
	WHERE book_id = $2
		AND author_id NOT IN (SELECT unnest($1::uuid[]));
`

// GetAuthorBooks
const getAuthorBooksQuery = `
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

// RegisterAuthor
const insertAuthorQuery = `
	INSERT INTO author (name)
	VALUES ($1)
	RETURNING id;
`

// GetAuthorInfo
const getAuthorQuery = `
	SELECT id, name
	FROM author
	WHERE id = $1;
`

// ChangeAuthor
const updateAuthorQuery = `
	UPDATE author SET name = $1 WHERE id = $2;
`

// Outbox
const markAsProcessedQuery = `
	UPDATE outbox
	SET status = 'SUCCESS'
	WHERE idempotency_key = ANY($1);
`

// Outbox
const getMessagesQuery = `
	UPDATE outbox
	SET status = 'IN_PROGRESS'
	WHERE idempotency_key IN (
    	SELECT idempotency_key
    	FROM outbox
		WHERE
        	(status = 'CREATED'
        		OR (status = 'IN_PROGRESS' AND updated_at < now() - $1::interval)) -- Явный каст времени к интервалу
    	ORDER BY created_at
    	LIMIT $2
    	FOR UPDATE SKIP LOCKED -- FIXME 
		)
	RETURNING idempotency_key, data, kind;
`

// Outbox
const sendMessageQuery = `
	INSERT INTO outbox (idempotency_key, data, status, kind)
	VALUES($1, $2, 'CREATED', $3)
	ON CONFLICT (idempotency_key) DO NOTHING -- Если уже существует, скип
`
