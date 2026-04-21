// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
//go:build integration

package db_test

import (
	"embed"
	"os"
	"testing"

	"eats/backend/common/testutils"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func TestMain(m *testing.M) {
	testutils.RunMigrations("orders", embedMigrations, "migrations")
	os.Exit(m.Run())
}
