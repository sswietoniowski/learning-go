package data

type Databaser interface {
	GetAll() []Book
	Add(book Book) Book
	GetById(id int64) (Book, bool)
	ModifyById(id int64, book Book) (Book, bool)
	RemoveById(id int64) (Book, bool)
}
