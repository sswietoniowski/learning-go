package data

import (
	"time"

	"github.com/jaswdr/faker/v2"
)

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
