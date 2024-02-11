package data

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq" // PostgreSQL driver
)

/* To start the PostgreSQL database as a Docker container, run the following command:

docker run --name readinglist -e POSTGRES_PASSWORD=PUT_REAL_PASSWORD_HERE -e POSTGRES_DB=readinglist -p 5433:5432 -d postgres

To setup the database, run the following commands in the terminal to copy the setup.sql file to the container and execute it with psql:

docker cp ./scripts/setup.sql readinglist:/setup.sql
docker exec -it readinglist psql -U postgres -d readinglist -f /setup.sql

To tear down the database, run the following commands in the terminal to copy the teardown.sql file to the container and execute it with psql:

docker cp ./scripts/teardown.sql readinglist:/teardown.sql
docker exec -it readinglist psql -U postgres -d readinglist -f /teardown.sql

*/

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

	selectAllQuery := `
SELECT id, title, author, published, pages, genres, rating, version, read, created_at
FROM books
`
	p.logger.Println(selectAllQuery)

	rows, err := db.Query(selectAllQuery)
	if err != nil {
		return nil, &DatabaseError{"GetAll", err}
	}

	books := make([]Book, 0) // to return an empty array instead of nil when there are no books

	for rows.Next() {
		var book Book
		var genres string
		err := rows.Scan(&book.Id, &book.Title, &book.Author, &book.Published,
			&book.Pages, &genres, &book.Rating, &book.Version, &book.Read, &book.CreatedAt)
		if err != nil {
			return nil, &DatabaseError{"GetAll", err}
		}
		book.Genres = convertPostgreSQLArrayToSlice(genres)
		books = append(books, book)
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

	insertQuery := `
INSERT INTO books (title, author, published, pages, genres, rating, version, read, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id
`
	p.logger.Println(insertQuery)

	genres := convertSliceToPostgreSQLArray(book.Genres)

	row := db.QueryRow(insertQuery,
		book.Title, book.Author, book.Published, book.Pages,
		genres, book.Rating, book.Version, book.Read, book.CreatedAt)

	err = row.Scan(&book.Id)
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

	selectByIdQuery := `
SELECT id, title, author, published, pages, genres, rating, version, read, created_at
FROM books
WHERE id = $1
`
	p.logger.Println(selectByIdQuery)

	row := db.QueryRow(selectByIdQuery, id)

	var book Book
	var genres string
	err = row.Scan(&book.Id, &book.Title, &book.Author, &book.Published,
		&book.Pages, &genres, &book.Rating, &book.Version, &book.Read, &book.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &NotFoundError{Id: id}
		}
		return nil, &DatabaseError{"GetById", err}
	}
	book.Genres = convertPostgreSQLArrayToSlice(genres)

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

	updateQuery := `
UPDATE books
SET title = $1, author = $2, published = $3, pages = $4, genres = $5, rating = $6,
	version = $7, read = $8, created_at = $9
WHERE id = $10
`
	p.logger.Println(updateQuery)

	genres := convertSliceToPostgreSQLArray(book.Genres)

	_, err = db.Exec(updateQuery, book.Title, book.Author, book.Published, book.Pages,
		genres, book.Rating, book.Version, book.Read, book.CreatedAt, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &NotFoundError{Id: id}
		}
		return nil, &DatabaseError{"ModifyById", err}
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

	deleteQuery := `
DELETE FROM books
WHERE id = $1
RETURNING id, title, author, published, pages, genres, rating, version, read, created_at
`
	p.logger.Println(deleteQuery)

	row := db.QueryRow(deleteQuery, id)

	var book Book
	var genres string
	err = row.Scan(&book.Id, &book.Title, &book.Author, &book.Published,
		&book.Pages, &genres, &book.Rating, &book.Version, &book.Read, &book.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &NotFoundError{Id: id}
		}
		return nil, &DatabaseError{"RemoveById", err}
	}
	book.Genres = convertPostgreSQLArrayToSlice(genres)

	return &book, nil
}

func convertSliceToPostgreSQLArray(slice []string) string {
	return fmt.Sprintf("{%s}", strings.Join(slice, ","))
}

func convertPostgreSQLArrayToSlice(postgreSQLArray string) []string {
	postgreSQLArray = postgreSQLArray[1 : len(postgreSQLArray)-1]
	return strings.Split(postgreSQLArray, ",")
}
