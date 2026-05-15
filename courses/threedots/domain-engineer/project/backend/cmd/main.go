package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ThreeDotsLabs/the-domain-engineer/clients"
	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend"
	"eats/backend/billing/adapters/tax"
	"eats/backend/common/file"
	"eats/backend/common/log"
	"eats/backend/orders/adapters/payments"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Init(slog.LevelInfo)

	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		panic("POSTGRES_URL environment variable is not set")
	}

	dbPgx, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}

	apiClients, err := clients.NewClientsWithHttpClient(
		os.Getenv("GATEWAY_ADDR"),
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Correlation-ID", log.CorrelationIDFromContext(ctx))
			return nil
		},
		&http.Client{Timeout: 10 * time.Second},
	)
	if err != nil {
		panic(err)
	}

	svc, err := backend.New(
		ctx,
		dbPgx,
		backend.ExternalServices{
			Payments:    payments.NewClient(apiClients),
			Tax:         tax.NewClient(apiClients),
			FileStorage: file.NewPublicStorage(apiClients),
		},
	)
	if err != nil {
		panic(err)
	}

	if err := svc.Run(ctx, ":8080"); err != nil {
		panic(err)
	}
}
