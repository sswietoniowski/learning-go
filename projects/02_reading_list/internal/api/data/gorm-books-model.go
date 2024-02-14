package data

import (
	"strings"
	"time"
)

// GormBook represents a book in the database using GORM.
type GormBook struct {
	Id        int64  `gorm:"primary_key"`
	Title     string `gorm:"type:text;not null"`
	Author    string `gorm:"type:text;not null"`
	Published int
	Pages     int
	Genres    string `gorm:"type:text"`
	Rating    float32
	Version   int       `gorm:"not null;default:1"`
	Read      bool      `gorm:"not null;default:false"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:current_timestamp()"`
}

// GormBookToBook converts a GormBook to a Book.
func GormBookToBook(gb *GormBook) *Book {
	return &Book{
		Id:        gb.Id,
		Title:     gb.Title,
		Author:    gb.Author,
		Published: gb.Published,
		Pages:     gb.Pages,
		Genres:    strings.Split(gb.Genres, ","),
		Rating:    gb.Rating,
		Version:   gb.Version,
		Read:      gb.Read,
		CreatedAt: gb.CreatedAt,
	}
}

// BookToGormBook converts a Book to a GormBook.
func BookToGormBook(b *Book) *GormBook {
	return &GormBook{
		Id:        b.Id,
		Title:     b.Title,
		Author:    b.Author,
		Published: b.Published,
		Pages:     b.Pages,
		Genres:    strings.Join(b.Genres, ","),
		Rating:    b.Rating,
		Version:   b.Version,
		Read:      b.Read,
		CreatedAt: b.CreatedAt,
	}
}

// GormBooksToBooks converts a slice of GormBooks to a slice of Books.
func GormBooksToBooks(gb []GormBook) []Book {
	books := make([]Book, 0, len(gb)) // pre-allocate the slice to avoid resizing it
	for _, b := range gb {
		books = append(books, *GormBookToBook(&b))
	}
	return books
}

// BooksToGormBooks converts a slice of Books to a slice of GormBooks.
func BooksToGormBooks(b []Book) []GormBook {
	books := make([]GormBook, 0, len(b)) // pre-allocate the slice to avoid resizing it
	for _, b := range b {
		books = append(books, *BookToGormBook(&b))
	}
	return books
}
