package data

import "fmt"

// NotFoundError is an error type for a record not found in the database.
type NotFoundError struct {
	// Id is the ID of the record that was not found.
	Id int64
}

// Error returns a string representation of the error.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Record with ID %d not found", e.Id)
}

// DatabaseError is an error type for a database operation error.
type DatabaseError struct {
	// Operation is the name of the operation that caused the error.
	Operation string
	// Err is the original error.
	Err error
}

// Error returns a string representation of the error.
func (e *DatabaseError) Error() string {
	return fmt.Sprintf("Database error during %s: %v", e.Operation, e.Err)
}

// Databaser is an interface for a database of books.
type Databaser interface {
	// GetAll returns all books from the database or an error if something went wrong.
	GetAll() ([]Book, error)
	// Add adds a new book to the database and returns the added book or an error if something went wrong.
	Add(book Book) (*Book, error)
	// GetById returns a book from the database by its id or an error if something went wrong.
	// If the book is not found, it returns NotFoundError as the error, for other errors it returns DatabaseError.
	GetById(id int64) (*Book, error)
	// ModifyById modifies a book in the database by its id and returns the modified book or an error if something went wrong.
	// If the book is not found, it returns NotFoundError as the error, for other errors it returns DatabaseError.
	ModifyById(id int64, book Book) (*Book, error)
	// RemoveById removes a book from the database by its id and returns the removed book or an error if something went wrong.
	// If the book is not found, it returns NotFoundError as the error, for other errors it returns DatabaseError.
	RemoveById(id int64) (*Book, error)
}
