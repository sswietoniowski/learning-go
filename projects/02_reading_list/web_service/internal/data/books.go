package data

import "time"

type Book struct {
	Id        int64     `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Published int       `json:"published"`
	Pages     int       `json:"pages"`
	Genres    []string  `json:"genres"`
	Rating    float32   `json:"rating"`
	Version   int       `json:"version"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

type Books struct {
	books []Book
}

func NewBooks() *Books {
	return &Books{
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

func (b *Books) All() []Book {
	return b.books
}

func (b *Books) Add(book Book) Book {
	book.Id = int64(len(b.books) + 1)
	book.CreatedAt = time.Now()
	b.books = append(b.books, book)
	return book
}

func (b *Books) FindById(id int64) (Book, bool) {
	for _, book := range b.books {
		if book.Id == id {
			return book, true
		}
	}
	return Book{}, false
}

func (b *Books) UpdateById(id int64, book Book) bool {
	for i, oldBook := range b.books {
		if oldBook.Id == id {
			book.Id = oldBook.Id
			book.CreatedAt = oldBook.CreatedAt
			b.books[i] = book
			return true
		}
	}
	return false
}

func (b *Books) DeleteById(id int64) bool {
	for i, book := range b.books {
		if book.Id == id {
			b.books = append(b.books[:i], b.books[i+1:]...)
			return true
		}
	}
	return false
}
