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
	Version   int       `json:"-"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"-"`
}
