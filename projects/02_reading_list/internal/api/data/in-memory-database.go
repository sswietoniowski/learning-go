package data

import (
	"log"
	"time"
)

// InMemoryDatabase is an in-memory database of books.
type InMemoryDatabase struct {
	books  []Book
	logger *log.Logger
}

// NewInMemoryDatabase creates a new in-memory database with some initial data.
func NewInMemoryDatabase(logger *log.Logger) *InMemoryDatabase {
	logger.Println("using in-memory database")

	var initialBooks = []Book{
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
	}

	return &InMemoryDatabase{
		books:  initialBooks,
		logger: logger,
	}
}

// GetAll returns all books from the database or an error if something went wrong.
func (b *InMemoryDatabase) GetAll() ([]Book, error) {
	b.logger.Println("get all books")

	return b.books, nil
}

// Add adds a new book to the database and returns the added book or an error if something went wrong.
func (b *InMemoryDatabase) Add(book Book) (*Book, error) {
	b.logger.Println("add book")

	book.Id = int64(len(b.books) + 1)
	book.CreatedAt = time.Now()
	b.books = append(b.books, book)

	return &book, nil
}

// GetById returns a book from the database by its id or an error if something went wrong.
// If the book is not found, it returns NotFoundError as the error, for other errors it returns DatabaseError.
func (b *InMemoryDatabase) GetById(id int64) (*Book, error) {
	b.logger.Println("get book by id")

	for _, book := range b.books {
		if book.Id == id {
			return &book, nil
		}
	}

	// book not found
	return nil, &NotFoundError{Id: id}
}

// ModifyById modifies a book in the database by its id and returns the modified book or an error if something went wrong.
// If the book is not found, it returns NotFoundError as the error, for other errors it returns DatabaseError.
func (b *InMemoryDatabase) ModifyById(id int64, book Book) (*Book, error) {
	b.logger.Println("modify book by id")

	for i, oldBook := range b.books {
		if oldBook.Id == id {
			book.Id = oldBook.Id
			book.CreatedAt = oldBook.CreatedAt
			b.books[i] = book

			return &book, nil
		}
	}

	// book not found
	return nil, &NotFoundError{Id: id}
}

// RemoveById removes a book from the database by its id and returns the removed book or an error if something went wrong.
// If the book is not found, it returns NotFoundError as the error, for other errors it returns DatabaseError.
func (b *InMemoryDatabase) RemoveById(id int64) (*Book, error) {
	b.logger.Println("remove book by id")

	for i, book := range b.books {
		if book.Id == id {
			b.books = append(b.books[:i], b.books[i+1:]...)

			return &book, nil
		}
	}

	// book not found
	return nil, &NotFoundError{Id: id}
}
