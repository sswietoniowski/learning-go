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

*/

// PostgreSQLDatabase is a database that uses PostgreSQL
type PostgreSQLDatabase struct {
	connectionString string
	logger           *log.Logger
}

// NewPostgreSQLDatabase creates a new PostgreSQLDatabase
func NewPostgreSQLDatabase(config PostgreSQLConfig, logger *log.Logger) *PostgreSQLDatabase {
	connectionString := config.ConnectionString()
	logger.Printf("connection string: %s\n", connectionString)

	return &PostgreSQLDatabase{
		connectionString: connectionString,
		logger:           logger,
	}
}

func (p *PostgreSQLDatabase) GetAll() []Book {
	p.logger.Println("get all books")
	db, err := sql.Open("postgres", p.connectionString)
	if err != nil {
		panic(err) // TODO: handle error
	}
	defer db.Close()

	selectAllQuery := `
SELECT id, title, author, published, pages, genres, rating, version, read, created_at
FROM books
`
	p.logger.Println(selectAllQuery)

	rows, err := db.Query(selectAllQuery)
	if err != nil {
		panic(err) // TODO: handle error
	}

	books := make([]Book, 0) // to return an empty JSON array instead of null

	for rows.Next() {
		var book Book
		var genres string
		err := rows.Scan(&book.Id, &book.Title, &book.Author, &book.Published,
			&book.Pages, &genres, &book.Rating, &book.Version, &book.Read, &book.CreatedAt)
		if err != nil {
			panic(err) // TODO: handle error
		}
		book.Genres = convertPostgreSQLArrayToSlice(genres)
		books = append(books, book)
	}

	return books
}

func (p *PostgreSQLDatabase) Add(book Book) Book {
	db, err := sql.Open("postgres", p.connectionString)
	if err != nil {
		panic(err) // TODO: handle error
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
		panic(err) // TODO: handle error
	}

	return book
}

func (p *PostgreSQLDatabase) GetById(id int64) (Book, bool) {
	db, err := sql.Open("postgres", p.connectionString)
	if err != nil {
		panic(err) // TODO: handle error
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
			return Book{}, false
		}
		panic(err) // TODO: handle error
	}
	book.Genres = convertPostgreSQLArrayToSlice(genres)

	return book, true
}

func (p *PostgreSQLDatabase) ModifyById(id int64, book Book) (Book, bool) {
	db, err := sql.Open("postgres", p.connectionString)
	if err != nil {
		panic(err) // TODO: handle error
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
			return Book{}, false
		}
		panic(err) // TODO: handle error
	}

	return book, true
}

func (p *PostgreSQLDatabase) RemoveById(id int64) (Book, bool) {
	db, err := sql.Open("postgres", p.connectionString)
	if err != nil {
		panic(err) // TODO: handle error
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
			return Book{}, false
		}
		panic(err) // TODO: handle error
	}
	book.Genres = convertPostgreSQLArrayToSlice(genres)

	return book, true
}

func convertSliceToPostgreSQLArray(slice []string) string {
	return fmt.Sprintf("{%s}", strings.Join(slice, ","))
}

func convertPostgreSQLArrayToSlice(postgreSQLArray string) []string {
	postgreSQLArray = postgreSQLArray[1 : len(postgreSQLArray)-1]
	return strings.Split(postgreSQLArray, ",")
}