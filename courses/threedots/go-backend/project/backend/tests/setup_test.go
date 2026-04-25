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
	commonHTTP "eats/backend/common/http"
	"eats/backend/common/log"
	ordersclient "eats/backend/orders/api/http/client"
)

type testClients struct {
	Orders *ordersclient.ClientWithResponses
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

	return testClients{
		Orders: orders,
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

	svc, err := backend.New(
		ctx,
		dbPgx,
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
