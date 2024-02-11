CREATE TABLE books (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    published INT,
    pages INT,
    genres TEXT[],
    rating REAL,
    version INT,
    read BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Initial books
INSERT INTO books
(id, title, author, published, pages, genres, rating, version, "read", created_at)
VALUES
(1, 'The Hitchhiker''s Guide to the Galaxy', 'Douglas Adams', 1979, 224, ARRAY['comedy', 'science fiction'], 5.0, 1, false, '2024-01-01 08:00:00'),
(2, 'The Hobbit', 'J.R.R. Tolkien', 1937, 310, ARRAY['adventure', 'fantasy'], 4.5, 1, true, '2024-01-01 09:00:00');

SELECT setval('books_id_seq', (SELECT MAX(id) FROM books));