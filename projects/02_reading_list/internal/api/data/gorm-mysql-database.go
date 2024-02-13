/*
To start the MySQL database as a Docker container, run the following command:

docker run --name readinglist -e MYSQL_ROOT_PASSWORD=PUT_REAL_PASSWORD_HERE -e MYSQL_DATABASE=readinglist -p 3306:3306 -d mysql
*/

package data

import (
	"log"

	_ "gorm.io/driver/mysql"
	_ "gorm.io/gorm"
)

type GormMySQLDatabase struct {
	dsn    string
	logger *log.Logger
}

// NewGormMySQLDatabase creates a new GormMySQLDatabase and returns it with the given DSN and logger.
func NewGormMySQLDatabase(dsn string, logger *log.Logger) *GormMySQLDatabase {
	logger.Println("using gorm-mysql database")

	// This is not a secure way to log the connection string, but it's useful for debugging and learning.
	// Don't do this in a production environment and for the real connection string.
	logger.Printf("dsn: %s\n", dsn)

	return &GormMySQLDatabase{
		dsn:    dsn,
		logger: logger,
	}
}

// GetAll returns all books from the database or an error if something went wrong.
func (g *GormMySQLDatabase) GetAll() ([]Book, error) {
	g.logger.Println("get all books")

	// TODO: Implement this method

	return nil, nil
}

// Add adds a new book to the database and returns the added book or an error if something went wrong.
func (g *GormMySQLDatabase) Add(book Book) (*Book, error) {
	g.logger.Println("add a new book")

	// TODO: Implement this method

	return nil, nil
}

// GetById returns a book from the database by its id or an error if something went wrong.
// If the book is not found, it returns NotFoundError as the error, for other errors it returns DatabaseError.
func (g *GormMySQLDatabase) GetById(id int64) (*Book, error) {
	g.logger.Println("get a book by id")

	// TODO: Implement this method

	return nil, nil
}

// ModifyById modifies a book in the database by its id and returns the modified book or an error if something went wrong.
// If the book is not found, it returns NotFoundError as the error, for other errors it returns DatabaseError.
func (g *GormMySQLDatabase) ModifyById(id int64, book Book) (*Book, error) {
	g.logger.Println("modify a book by id")

	// TODO: Implement this method

	return nil, nil
}

// RemoveById removes a book from the database by its id and returns the removed book or an error if something went wrong.
// If the book is not found, it returns NotFoundError as the error, for other errors it returns DatabaseError.
func (g *GormMySQLDatabase) RemoveById(id int64) (*Book, error) {
	g.logger.Println("remove a book by id")

	// TODO: Implement this method

	return nil, nil
}
