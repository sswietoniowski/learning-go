/*
To start the PostgreSQL database as a Docker container, run the following command:

docker run --name readinglist -e POSTGRES_PASSWORD=PUT_REAL_PASSWORD_HERE -e POSTGRES_DB=readinglist -p 5433:5432 -d postgres

To setup the database, run the following commands in the terminal to copy the setup.sql file to the container and execute it with psql:

docker cp ./scripts/setup.sql readinglist:/setup.sql
docker exec -it readinglist psql -U postgres -d readinglist -f /setup.sql

To tear down the database, run the following commands in the terminal to copy the teardown.sql file to the container and execute it with psql:

docker cp ./scripts/teardown.sql readinglist:/teardown.sql
docker exec -it readinglist psql -U postgres -d readinglist -f /teardown.sql
*/

package data

import (
	"database/sql"
	"log"

	"github.com/lib/pq" // PostgreSQL driver
)

// PostgreSQLDatabase is a PostgreSQL database of books.
type PostgreSQLDatabase struct {
	dsn    string
	logger *log.Logger
}

// NewPostgreSQLDatabase creates a new PostgreSQLDatabase with the given DSN and logger.
func NewPostgreSQLDatabase(dsn string, logger *log.Logger) *PostgreSQLDatabase {
	logger.Println("using postgresql database")

	// This is not a secure way to log the connection string, but it's useful for debugging and learning.
	// Don't do this in a production environment and for the real connection string.
	logger.Printf("dsn: %s\n", dsn)

	return &PostgreSQLDatabase{
		dsn:    dsn,
		logger: logger,
	}
}

// GetAll returns all books from the database or an error if something went wrong.
func (p *PostgreSQLDatabase) GetAll() ([]Book, error) {
	p.logger.Println("get all books")

	db, err := sql.Open("postgres", p.dsn)
	if err != nil {
		return nil, &DatabaseError{"GetAll", err}
	}
	defer db.Close()

	query := `
SELECT id, title, author, published, pages, genres, rating, version, read, created_at
FROM books
ORDER BY id ASC
`
	rows, err := db.Query(query)
	if err != nil {
		return nil, &DatabaseError{"GetAll", err}
	}
	defer rows.Close()

	books := make([]Book, 0) // to return an empty array instead of nil when there are no books

	for rows.Next() {
		var book Book

		err := rows.Scan(
			&book.Id,
			&book.Title,
			&book.Author,
			&book.Published,
			&book.Pages,
			pq.Array(&book.Genres),
			&book.Rating,
			&book.Version,
			&book.Read,
			&book.CreatedAt,
		)
		if err != nil {
			return nil, &DatabaseError{"GetAll", err}
		}

		books = append(books, book)
	}

	if err = rows.Err(); err != nil {
		return nil, &DatabaseError{"GetAll", err}
	}

	return books, nil
}

// Add adds a new book to the database and returns the added book or an error if something went wrong.
func (p *PostgreSQLDatabase) Add(book Book) (*Book, error) {
	db, err := sql.Open("postgres", p.dsn)
	if err != nil {
		return nil, &DatabaseError{"Add", err}
	}
	defer db.Close()

	query := `
INSERT INTO books (title, author, published, pages, genres, rating, read)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at, version
`
	args := []interface{}{
		book.Title,
		book.Author,
		book.Published,
		book.Pages,
		pq.Array(book.Genres),
		book.Rating,
		book.Read,
	}

	err = db.QueryRow(query, args...).Scan(
		&book.Id,
		&book.CreatedAt,
		&book.Version,
	)
	if err != nil {
		return nil, &DatabaseError{"Add", err}
	}

	return &book, nil
}

// GetById returns a book from the database by its id or an error if something went wrong.
// If the book is not found, it returns NotFoundError as the error, for other errors it returns DatabaseError.
func (p *PostgreSQLDatabase) GetById(id int64) (*Book, error) {
	db, err := sql.Open("postgres", p.dsn)
	if err != nil {
		return nil, &DatabaseError{"GetById", err}
	}
	defer db.Close()

	query := `
SELECT id, title, author, published, pages, genres, rating, version, read, created_at
FROM books
WHERE id = $1
`
	var book Book
	err = db.QueryRow(query, id).Scan(
		&book.Id,
		&book.Title,
		&book.Author,
		&book.Published,
		&book.Pages,
		pq.Array(&book.Genres),
		&book.Rating,
		&book.Version,
		&book.Read,
		&book.CreatedAt,
	)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, &NotFoundError{Id: id}
		default:
			return nil, &DatabaseError{"GetById", err}
		}
	}

	return &book, nil
}

// ModifyById modifies a book in the database by its id and returns the modified book or an error if something went wrong.
// If the book is not found, it returns NotFoundError as the error, for other errors it returns DatabaseError.
func (p *PostgreSQLDatabase) ModifyById(id int64, book Book) (*Book, error) {
	db, err := sql.Open("postgres", p.dsn)
	if err != nil {
		return nil, &DatabaseError{"ModifyById", err}
	}
	defer db.Close()

	query := `
UPDATE books
SET title = $1, author = $2, published = $3, pages = $4, genres = $5, rating = $6, version = version + 1, read = $7
WHERE id = $8
RETURNING version
`
	args := []interface{}{
		book.Title,
		book.Author,
		book.Published,
		book.Pages,
		pq.Array(book.Genres),
		book.Rating,
		book.Read,
		id,
	}

	err = db.QueryRow(query, args...).Scan(&book.Version)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, &NotFoundError{Id: id}
		default:
			return nil, &DatabaseError{"ModifyById", err}
		}
	}

	return &book, nil
}

// RemoveById removes a book from the database by its id and returns the removed book or an error if something went wrong.
// If the book is not found, it returns NotFoundError as the error, for other errors it returns DatabaseError.
func (p *PostgreSQLDatabase) RemoveById(id int64) (*Book, error) {
	db, err := sql.Open("postgres", p.dsn)
	if err != nil {
		return nil, &DatabaseError{"RemoveById", err}
	}
	defer db.Close()

	query := `
DELETE FROM books
WHERE id = $1
RETURNING id, title, author, published, pages, genres, rating, version, read, created_at
`
	var book Book
	err = db.QueryRow(query, id).Scan(
		&book.Id,
		&book.Title,
		&book.Author,
		&book.Published,
		&book.Pages,
		pq.Array(&book.Genres),
		&book.Rating,
		&book.Version,
		&book.Read,
		&book.CreatedAt,
	)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, &NotFoundError{Id: id}
		default:
			return nil, &DatabaseError{"RemoveById", err}
		}
	}

	return &book, nil
}
