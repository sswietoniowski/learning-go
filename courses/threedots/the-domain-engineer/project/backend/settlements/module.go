package settlements

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/common"
	"eats/backend/common/module"
	"eats/backend/common/module/contracts"
	"eats/backend/settlements/adapters/db"
	"eats/backend/settlements/api/http"
	settlementsModule "eats/backend/settlements/api/module"
	"eats/backend/settlements/app/command"
	"eats/backend/settlements/app/query"
)

type Module struct {
	pgxDb *pgxpool.Pool

	modules *contracts.Contracts

	commandHandlers       *command.Handlers
	queryHandlers         *query.Handlers
	legalEntityRepository *db.LegalEntityRepository
}

func NewModule(pgxDb *pgxpool.Pool, modules *contracts.Contracts) *Module {
	return &Module{
		pgxDb:   pgxDb,
		modules: modules,
	}
}

func (m *Module) Name() module.Name {
	return "settlements"
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

	billingCycleRepository := db.NewBillingCycleRepository(m.pgxDb)
	orderRepository := db.NewOrderRepository(m.pgxDb)
	m.legalEntityRepository = db.NewLegalEntityRepository(m.pgxDb)

	m.commandHandlers = command.NewHandlers(
		billingCycleRepository,
		orderRepository,
		m.legalEntityRepository,
		m.modules,
	)

	m.queryHandlers = query.NewHandlers(billingCycleRepository)

	return nil
}

func (m *Module) RegisterContracts(ctx context.Context, contracts *contracts.Contracts) error {
	contracts.Settlements = settlementsModule.New(m.commandHandlers, m.legalEntityRepository)
	return nil
}

func (m *Module) RegisterHttp(ctx context.Context, e common.EchoRouter) error {
	return http.Register(e, m.commandHandlers, m.queryHandlers)
}
