package billing

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/common"
	"eats/backend/common/module"
	"eats/backend/common/module/contracts"
)

type Module struct {
	pgxDb *pgxpool.Pool
}

func NewModule(pgxDb *pgxpool.Pool) *Module {
	return &Module{pgxDb: pgxDb}
}

func (m *Module) Name() module.Name {
	return "billing"
}

//go:embed adapters/db/migrations/*.sql
var embedMigrations embed.FS

func (m *Module) Init(ctx context.Context) error {
	return common.MigrateDatabaseUp(
		ctx,
		string(m.Name()),
		m.pgxDb,
		embedMigrations,
		"adapters/db/migrations",
	)
}

func (m *Module) RegisterContracts(ctx context.Context, contracts *contracts.Contracts) error {
	return nil
}

func (m *Module) RegisterHttp(ctx context.Context, e common.EchoRouter) error {
	return nil
}
