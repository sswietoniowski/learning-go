package testutils

import (
	"context"
	"io/fs"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"eats/backend/common"
)

// RunMigrations runs database migrations before tests execute.
// Call this from TestMain in repository test packages.
// embedFS must contain migration files, typically via go:embed directive.
func RunMigrations(moduleName string, embedFS fs.FS, migrationsDir string) {
	ctx := context.Background()

	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		panic("POSTGRES_URL environment variable is not set")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	if err := common.MigrateDatabaseUp(ctx, moduleName, pool, embedFS, migrationsDir); err != nil {
		panic(err)
	}
}

func NewDB(t *testing.T) *pgxpool.Pool {
	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		panic("POSTGRES_URL environment variable is not set")
	}

	dbPgx, err := pgxpool.New(t.Context(), dsn)
	require.NoError(t, err)

	return dbPgx
}
