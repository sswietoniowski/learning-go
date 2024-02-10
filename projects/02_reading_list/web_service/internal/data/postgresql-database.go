package data

import "database/sql"

type PostgreSQLDatabase struct {
	DbConfig DbConfig
}

func NewPostgreSQLDatabase(config DbConfig) *PostgreSQLDatabase {
	return &PostgreSQLDatabase{
		DbConfig: config,
	}
}

func (p *PostgreSQLDatabase) GetAll() []Book {
	db, err := sql.Open("postgres", p.DbConfig.ConnectionString())
	if err != nil {
		panic(err) // TODO: handle error
	}
	defer db.Close()

	// select * from books
	rows, err := db.Query(`
SELECT id, title, author, published, pages, genres, rating, version, read, created_at 
FROM books
	`)
	if err != nil {
		panic(err) // TODO: handle error
	}

	var books []Book

	for rows.Next() {
		var book Book
		err := rows.Scan(&book.Id, &book.Title, &book.Author, &book.Published,
			&book.Pages, &book.Genres, &book.Rating, &book.Version, &book.Read, &book.CreatedAt)
		if err != nil {
			panic(err) // TODO: handle error
		}
		books = append(books, book)
	}

	return books
}

func (p *PostgreSQLDatabase) Add(book Book) Book {
	db, err := sql.Open("postgres", p.DbConfig.ConnectionString())
	if err != nil {
		panic(err) // TODO: handle error
	}
	defer db.Close()

	_, err = db.Exec(`
INSERT INTO books (title, author, published, pages, genres, rating, version, read, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, book.Title, book.Author, book.Published, book.Pages,
		book.Genres, book.Rating, book.Version, book.Read, book.CreatedAt)

	if err != nil {
		panic(err) // TODO: handle error
	}

	return book
}

func (p *PostgreSQLDatabase) GetById(id int64) (Book, bool) {
	db, err := sql.Open("postgres", p.DbConfig.ConnectionString())
	if err != nil {
		panic(err) // TODO: handle error
	}
	defer db.Close()

	row := db.QueryRow(`
SELECT id, title, author, published, pages, genres, rating, version, read, created_at
FROM books
WHERE id = $1
		`, id)

	var book Book
	err = row.Scan(&book.Id, &book.Title, &book.Author, &book.Published,
		&book.Pages, &book.Genres, &book.Rating, &book.Version, &book.Read, &book.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Book{}, false
		}
		panic(err) // TODO: handle error
	}

	return book, true
}

func (p *PostgreSQLDatabase) ModifyById(id int64, book Book) (Book, bool) {
	db, err := sql.Open("postgres", p.DbConfig.ConnectionString())
	if err != nil {
		panic(err) // TODO: handle error
	}
	defer db.Close()

	_, err = db.Exec(`
UPDATE books
SET
	title = $1, author = $2, published = $3, pages = $4, genres = $5, rating = $6,
	version = $7, read = $8, created_at = $9
WHERE id = $10
		`, book.Title, book.Author, book.Published, book.Pages,
		book.Genres, book.Rating, book.Version, book.Read, book.CreatedAt, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return Book{}, false
		}
		panic(err) // TODO: handle error
	}

	return book, true
}

func (p *PostgreSQLDatabase) RemoveById(id int64) (Book, bool) {
	db, err := sql.Open("postgres", p.DbConfig.ConnectionString())
	if err != nil {
		panic(err) // TODO: handle error
	}
	defer db.Close()

	row := db.QueryRow(`
DELETE FROM books
WHERE id = $1
RETURNING id, title, author, published, pages, genres, rating, version, read, created_at
	`, id)

	var book Book
	err = row.Scan(&book.Id, &book.Title, &book.Author, &book.Published,
		&book.Pages, &book.Genres, &book.Rating, &book.Version, &book.Read, &book.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Book{}, false
		}
		panic(err) // TODO: handle error
	}

	return book, true
}
