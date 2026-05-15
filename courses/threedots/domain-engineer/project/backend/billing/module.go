package billing

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"

	billingdb "eats/backend/billing/adapters/db"
	"eats/backend/billing/adapters/printer"
	"eats/backend/billing/api/http"
	billingModule "eats/backend/billing/api/module"
	"eats/backend/billing/app/command"
	"eats/backend/billing/app/query"
	"eats/backend/common"
	"eats/backend/common/module"
	"eats/backend/common/module/contracts"
)

type fileStorage interface {
	StoreFile(ctx context.Context, path string, content []byte) (string, error)
}

type Module struct {
	pgxDb *pgxpool.Pool

	commandHandlers *command.Handlers
	queryHandlers   *query.Handlers

	fileStorage fileStorage
}

func NewModule(pgxDb *pgxpool.Pool, fileStorage fileStorage) *Module {
	return &Module{
		pgxDb:       pgxDb,
		fileStorage: fileStorage,
	}
}

func (m *Module) Name() module.Name {
	return "billing"
}

//go:embed adapters/db/migrations/*.sql
var embedMigrations embed.FS

func (m *Module) Init(ctx context.Context) error {
	if err := common.MigrateDatabaseUp(
		ctx,
		string(m.Name()),
		m.pgxDb,
		embedMigrations,
		"adapters/db/migrations",
	); err != nil {
		return err
	}

	documentPrinter := printer.NewPrinter()
	postgresRepo := billingdb.NewPostgresRepository(m.pgxDb)

	m.commandHandlers = command.NewHandlers(postgresRepo, documentPrinter, m.fileStorage)
	m.queryHandlers = query.NewHandlers(postgresRepo)

	return nil
}

func (m *Module) RegisterContracts(ctx context.Context, contracts *contracts.Contracts) error {
	contracts.Billing = billingModule.New(m.commandHandlers)
	return nil
}

func (m *Module) RegisterHttp(ctx context.Context, e common.EchoRouter) error {
	return http.Register(ctx, e, m.commandHandlers, m.queryHandlers)
}
