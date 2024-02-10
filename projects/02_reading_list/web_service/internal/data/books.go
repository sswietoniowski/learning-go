package data

import "time"

// Book represents a book in the database.
type Book struct {
	Id        int64     `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Published int       `json:"year,omitempty"`
	Pages     int       `json:"pages,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Rating    float32   `json:"rating,omitempty"`
	Version   int       `json:"version,omitempty"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"-"`
}

// Database is an in-memory database of books.
type Database struct {
	books []Book
}

// NewDatabase creates a new Database with some initial data.
func NewDatabase() *Database {
	return &Database{
		books: []Book{
			{
				Id:        1,
				Title:     "The Hitchhiker's Guide to the Galaxy",
				Author:    "Douglas Adams",
				Published: 1979,
				Pages:     224,
				Genres:    []string{"comedy", "science fiction"},
				Rating:    5.0,
				Version:   1,
				Read:      false,
				CreatedAt: time.Now(),
			},
			{
				Id:        2,
				Title:     "The Hobbit",
				Author:    "J.R.R. Tolkien",
				Published: 1937,
				Pages:     310,
				Genres:    []string{"adventure", "fantasy"},
				Rating:    4.5,
				Version:   1,
				Read:      true,
				CreatedAt: time.Now(),
			},
		},
	}
}

// GetAll returns all books from the database.
func (b *Database) GetAll() []Book {
	return b.books
}

// Add adds a new book to the database.
func (b *Database) Add(book Book) Book {
	book.Id = int64(len(b.books) + 1)
	book.CreatedAt = time.Now()
	b.books = append(b.books, book)
	return book
}

// GetById returns a book from the database by its id or false if not found.
func (b *Database) GetById(id int64) (Book, bool) {
	for _, book := range b.books {
		if book.Id == id {
			return book, true
		}
	}
	return Book{}, false
}

// ModifyById modifies a book in the database by its id or false if not found.
func (b *Database) ModifyById(id int64, book Book) (Book, bool) {
	for i, oldBook := range b.books {
		if oldBook.Id == id {
			book.Id = oldBook.Id
			book.CreatedAt = oldBook.CreatedAt
			b.books[i] = book
			return book, true
		}
	}
	return Book{}, false
}

// RemoveById removes a book from the database by its id or false if not found.
func (b *Database) RemoveById(id int64) (Book, bool) {
	for i, book := range b.books {
		if book.Id == id {
			b.books = append(b.books[:i], b.books[i+1:]...)
			return book, true
		}
	}
	return Book{}, false
}
