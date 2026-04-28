package db

import "github.com/jackc/pgx/v5/pgxpool"

type CourierRepository struct {
	db *pgxpool.Pool
}

func NewCourierRepository(db *pgxpool.Pool) *CourierRepository {
	return &CourierRepository{
		db: db,
	}
}
