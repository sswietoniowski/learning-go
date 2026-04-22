package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/orders/adapters/db/dbmodels"
	"eats/backend/orders/app"
)

type CustomerRepository struct {
	db *pgxpool.Pool
}

func NewCustomerRepository(db *pgxpool.Pool) *CustomerRepository {
	if db == nil {
		panic("db connection pool cannot be nil")
	}

	return &CustomerRepository{
		db: db,
	}
}

func (r *CustomerRepository) RegisterCustomer(ctx context.Context, customer app.Customer) error {
	queries := dbmodels.New(r.db)

	err := queries.InsertCustomer(ctx, dbmodels.InsertCustomerParams{
		CustomerUuid: customer.CustomerUUID,
		Name:         customer.Name,
		Email:        customer.Email,
		Address:      customer.Address,
		PhoneNumber:  customer.PhoneNumber,
	})
	if err != nil {
		return fmt.Errorf("insert customer failed: %w", err)
	}

	return nil
}
