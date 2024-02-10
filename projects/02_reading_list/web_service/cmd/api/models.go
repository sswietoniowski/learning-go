package main

type Book struct {
	Id      int64    `json:"id"`
	Title   string   `json:"title"`
	Author  string   `json:"author"`
	Year    int      `json:"year"`
	Pages   int      `json:"pages"`
	Genres  []string `json:"genres"`
	Rating  float32  `json:"rating"`
	Version int      `json:"version"`
	Read    bool     `json:"read"`
}

var books = []Book{
	{
		Id:      1,
		Title:   "The Hitchhiker's Guide to the Galaxy",
		Author:  "Douglas Adams",
		Year:    1979,
		Pages:   224,
		Genres:  []string{"comedy", "science fiction"},
		Rating:  5.0,
		Version: 1,
		Read:    false,
	},
	{
		Id:      2,
		Title:   "The Hobbit",
		Author:  "J.R.R. Tolkien",
		Year:    1937,
		Pages:   310,
		Genres:  []string{"adventure", "fantasy"},
		Rating:  4.5,
		Version: 1,
		Read:    true,
	},
}
