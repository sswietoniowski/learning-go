// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package tests_test

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend"
	"eats/backend/billing/adapters/tax"
	billingclient "eats/backend/billing/api/http/client"
	commonHTTP "eats/backend/common/http"
	"eats/backend/common/log"
	"eats/backend/orders/adapters/payments"
	ordersclient "eats/backend/orders/api/http/client"
	settlementclient "eats/backend/settlements/api/http/client"
)

type testStubs struct {
	Payments *payments.StubClient
	Tax      *tax.StubClient
}

var stubs testStubs

type testClients struct {
	Orders      *ordersclient.ClientWithResponses
	Billing     *billingclient.ClientWithResponses
	Settlements *settlementclient.ClientWithResponses
}

func newTestClients(t *testing.T) testClients {
	t.Helper()

	httpClient := &http.Client{Timeout: 10 * time.Second}

	editorFn := func(ctx context.Context, req *http.Request) error {
		log.FromContext(ctx).
			With(
				"method", req.Method,
				"url", req.URL.String(),
				"test_name", t.Name(),
			).
			Info("Making component test API request")
		req.Header.Set(commonHTTP.TestNameHeader, t.Name())
		return nil
	}

	orders, err := ordersclient.NewClientWithResponses("http://localhost:9090/",
		ordersclient.WithHTTPClient(httpClient),
		ordersclient.WithRequestEditorFn(editorFn),
	)
	if err != nil {
		t.Fatalf("creating orders client: %v", err)
	}

	billing, err := billingclient.NewClientWithResponses("http://localhost:9090/",
		billingclient.WithHTTPClient(httpClient),
		billingclient.WithRequestEditorFn(editorFn),
	)
	if err != nil {
		t.Fatalf("creating billing client: %v", err)
	}

	settlements, err := settlementclient.NewClientWithResponses("http://localhost:9090/",
		settlementclient.WithHTTPClient(httpClient),
		settlementclient.WithRequestEditorFn(editorFn),
	)
	if err != nil {
		t.Fatalf("creating settlements client: %v", err)
	}

	return testClients{
		Orders:      orders,
		Billing:     billing,
		Settlements: settlements,
	}
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
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

	paymentsStub := payments.NewStub()
	taxStub := tax.NewStub()

	svc, err := backend.New(
		ctx,
		dbPgx,
		backend.ExternalServices{
			Payments: paymentsStub,
			Tax:      taxStub,
		},
	)
	if err != nil {
		panic(err)
	}

	go func() {
		if err := svc.Run(ctx, ":9090"); err != nil {
			panic(err)
		}
	}()

	waitForHttpServerInMain()

	stubs = testStubs{
		Payments: paymentsStub,
		Tax:      taxStub,
	}

	exitCode := m.Run()

	cancel()

	os.Exit(exitCode)
}

func waitForHttpServerInMain() {
	maxAttempts := 100
	for i := 0; i < maxAttempts; i++ {
		resp, err := http.Get("http://localhost:9090/health")
		if err == nil && resp.StatusCode < 300 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(50 * time.Millisecond)
	}
	panic("HTTP server did not start in time")
}
