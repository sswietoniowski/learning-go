/*
To start the MySQL database as a Docker container, run the following command:

docker run --name readinglist -e MYSQL_ROOT_PASSWORD=PUT_REAL_PASSWORD_HERE -e MYSQL_DATABASE=readinglist -p 3307:3306 -d mysql
*/

package data

import (
	"log"
	"time"

	"github.com/jaswdr/faker/v2"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type GormMySQLDatabase struct {
	dsn    string
	logger *log.Logger
	db     *gorm.DB
}

func generateBooks(count int) []Book {
	var books []Book
	fake := faker.New()
	for i := 0; i < count; i++ {
		book := Book{
			Id:        int64(i + 3),
			Title:     fake.Lorem().Sentence(3),
			Author:    fake.Person().Name(),
			Published: fake.IntBetween(1900, 2021),
			Pages:     fake.IntBetween(100, 1000),
			Genres: []string{
				fake.Lorem().Word(),
			},
			Rating:    fake.Float32(1, 0, 5),
			Version:   1,
			Read:      false,
			CreatedAt: time.Now(),
		}
		books = append(books, book)
	}
	return books
}

// NewGormMySQLDatabase creates a new GormMySQLDatabase and returns it with the given DSN and logger.
func NewGormMySQLDatabase(dsn string, logger *log.Logger) *GormMySQLDatabase {
	logger.Println("using gorm-mysql database")

	// This is not a secure way to log the connection string, but it's useful for debugging and learning.
	// Don't do this in a production environment and for the real connection string.
	logger.Printf("dsn: %s\n", dsn)

	var g = &GormMySQLDatabase{
		dsn:    dsn,
		logger: logger,
	}

	err := g.open()
	if err != nil {
		logger.Printf("error: %s\n", err)
	}

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

	// just for fun use faker to generate a lot of books
	const booksQuantity = 100
	fakeBooks := generateBooks(booksQuantity)
	logger.Printf("generated %d books\n", booksQuantity)

	initialBooks = append(initialBooks, fakeBooks...)

	gormBooks := BooksToGormBooks(initialBooks)

	logger.Println("migrate the schema")

	err = g.db.AutoMigrate(&GormBook{})
	if err != nil {
		logger.Printf("error: %s\n", err)
	}

	logger.Println("insert initial books if the table is empty")

	var count int64
	g.db.Model(&GormBook{}).Count(&count)
	if count == 0 {
		logger.Println("insert initial books")

		for _, book := range gormBooks {
			result := g.db.Create(&book)
			if result.Error != nil {
				logger.Printf("error: %s\n", result.Error)
			}
		}
	}

	return g
}

func (g *GormMySQLDatabase) open() error {
	g.logger.Println("open the database")

	d, err := gorm.Open(mysql.Open(g.dsn), &gorm.Config{})
	if err != nil {
		return &DatabaseError{"open", err}
	}

	g.db = d

	return nil
}

func (g *GormMySQLDatabase) close() {
	g.logger.Println("close the database")

	if g.db == nil {
		return
	}

	sqlDB, err := g.db.DB()
	if err != nil {
		g.logger.Printf("error: %s\n", err)
	}

	err = sqlDB.Close()
	if err != nil {
		g.logger.Printf("error: %s\n", err)
	}
}

// GetAll returns all books from the database or an error if something went wrong.
func (g *GormMySQLDatabase) GetAll() ([]Book, error) {
	g.logger.Println("get all books")

	books := make([]GormBook, 0) // empty slice to handle the case when there are no books in the database
	result := g.db.Find(&books)
	if result.Error != nil {
		return nil, &DatabaseError{"GetAll", result.Error}
	}

	bks := GormBooksToBooks(books)

	return bks, nil
}

// Add adds a new book to the database and returns the added book or an error if something went wrong.
func (g *GormMySQLDatabase) Add(book Book) (*Book, error) {
	g.logger.Println("add a new book")

	gb := BookToGormBook(&book)

	result := g.db.Create(gb)
	if result.Error != nil {
		return nil, &DatabaseError{"Add", result.Error}
	}

	bk := GormBookToBook(gb)

	return bk, nil
}

// GetById returns a book from the database by its id or an error if something went wrong.
// If the book is not found, it returns NotFoundError as the error, for other errors it returns DatabaseError.
func (g *GormMySQLDatabase) GetById(id int64) (*Book, error) {
	g.logger.Println("get a book by id")

	var book GormBook
	result := g.db.First(&book, id)
	if result.Error != nil {
		switch result.Error {
		case gorm.ErrRecordNotFound:
			return nil, &NotFoundError{Id: id}
		default:
			return nil, &DatabaseError{"GetById", result.Error}
		}
	}

	bk := GormBookToBook(&book)

	return bk, nil
}

// ModifyById modifies a book in the database by its id and returns the modified book or an error if something went wrong.
// If the book is not found, it returns NotFoundError as the error, for other errors it returns DatabaseError.
func (g *GormMySQLDatabase) ModifyById(id int64, book Book) (*Book, error) {
	g.logger.Println("modify a book by id")

	bookFromDb, err := g.GetById(id)
	if err != nil {
		return nil, err
	}

	gb := BookToGormBook(&book)

	gb.Id = id
	gb.Version = bookFromDb.Version + 1
	gb.CreatedAt = bookFromDb.CreatedAt
	result := g.db.Save(gb)
	if result.Error != nil {
		switch result.Error {
		case gorm.ErrRecordNotFound:
			return nil, &NotFoundError{Id: id}
		default:
			return nil, &DatabaseError{"ModifyById", result.Error}
		}
	}

	bk := GormBookToBook(gb)

	return bk, nil
}

// RemoveById removes a book from the database by its id and returns the removed book or an error if something went wrong.
// If the book is not found, it returns NotFoundError as the error, for other errors it returns DatabaseError.
func (g *GormMySQLDatabase) RemoveById(id int64) (*Book, error) {
	g.logger.Println("remove a book by id")

	var book GormBook
	result := g.db.First(&book, id)
	if result.Error != nil {
		switch result.Error {
		case gorm.ErrRecordNotFound:
			return nil, &NotFoundError{Id: id}
		default:
			return nil, &DatabaseError{"RemoveById", result.Error}
		}
	}

	result = g.db.Delete(&book, id)
	if result.Error != nil {
		return nil, &DatabaseError{"RemoveById", result.Error}
	}

	bk := GormBookToBook(&book)

	return bk, nil
}
