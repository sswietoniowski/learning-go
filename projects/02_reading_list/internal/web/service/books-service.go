package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type BooksService struct {
	backendEndpoint string
	logger          *log.Logger
}

const baseUrl = "http://localhost:4000/api/v1"

var booksApiUrl = fmt.Sprintf("%s/books", baseUrl)

func NewBookService(backendEndpoint string, logger *log.Logger) *BooksService {
	return &BooksService{
		backendEndpoint: backendEndpoint,
		logger:          logger,
	}
}

func (s *BooksService) GetAll() (*[]Book, error) {
	url := fmt.Sprintf("%s/books", s.backendEndpoint)
	resp, err := http.Get(url)
	if err != nil {
		s.logger.Printf("error: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Printf("unexpected status: %s", resp.Status)
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Printf("error: %v\n", err)
		return nil, err
	}

	var books *[]Book
	err = json.Unmarshal(data, &books)
	if err != nil {
		s.logger.Printf("error: %v\n", err)
		return nil, err
	}

	return books, nil
}

func (s *BooksService) Get(id int64) (*Book, error) {
	url := fmt.Sprintf("%s/books/%d", s.backendEndpoint, id)

	resp, err := http.Get(url)
	if err != nil {
		s.logger.Printf("error: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Printf("unexpected status: %s", resp.Status)
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Printf("error: %v\n", err)
		return nil, err
	}

	var book *Book
	err = json.Unmarshal(data, &book)
	if err != nil {
		s.logger.Printf("error: %v\n", err)
		return nil, err
	}

	return book, nil
}
