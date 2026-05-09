package common

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	backoff "github.com/cenkalti/backoff/v5"
	"github.com/jackc/pgerrcode"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const couldNotSerializeAccessErrMsg = "could not serialize access"

type Beginner interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

func UpdateInTx(
	ctx context.Context,
	db Beginner,
	fn func(ctx context.Context, tx pgx.Tx) error,
) error {
	return updateInTxWithIsolation(ctx, db, pgx.RepeatableRead, fn)
}

func UpdateInReadCommittedTx(
	ctx context.Context,
	db Beginner,
	fn func(ctx context.Context, tx pgx.Tx) error,
) error {
	return updateInTxWithIsolation(ctx, db, pgx.ReadCommitted, fn)
}

func updateInTxWithIsolation(
	ctx context.Context,
	db Beginner,
	isoLevel pgx.TxIsoLevel,
	fn func(ctx context.Context, tx pgx.Tx) error,
) error {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = time.Millisecond
	b.MaxInterval = 500 * time.Millisecond
	b.Multiplier = 2.0
	b.RandomizationFactor = 0.5

	_, err := backoff.Retry(
		ctx,
		func() (struct{}, error) {
			err := updateInTx(ctx, db, isoLevel, fn)
			if err != nil {
				if strings.Contains(err.Error(), couldNotSerializeAccessErrMsg) {
					// Retryable
					return struct{}{}, err
				} else {
					return struct{}{}, backoff.Permanent(err)
				}
			}

			return struct{}{}, nil
		},
		backoff.WithBackOff(b),
		backoff.WithMaxTries(10),
	)

	return err
}

func updateInTx(
	ctx context.Context,
	db Beginner,
	isoLevel pgx.TxIsoLevel,
	fn func(ctx context.Context, tx pgx.Tx) error,
) (err error) {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: isoLevel})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("could not begin transaction (possible pool exhaustion — context deadline exceeded): %w", err)
		}
		return fmt.Errorf("could not begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = errors.Join(err, rollbackErr)
			}
			return
		}

		err = tx.Commit(ctx)
	}()

	return fn(ctx, tx)
}

func IsUniqueViolationError(err error, constraint string) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == constraint
	}

	return false
}
