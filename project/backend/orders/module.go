package orders

import (
	"context"
	"embed"

	"github.com/ThreeDotsLabs/the-domain-engineer/clients"
	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/common"
	"eats/backend/common/module"
	"eats/backend/common/module/contracts"
	"eats/backend/orders/adapters/db"
	"eats/backend/orders/adapters/payments"
	http2 "eats/backend/orders/api/http"
	ordersModule "eats/backend/orders/api/module"
	"eats/backend/orders/app"
)

type Module struct {
	pgxDb       *pgxpool.Pool
	httpHandler http2.Handler

	modules *contracts.Contracts
	clients *clients.Clients
}

func NewModule(pgxDb *pgxpool.Pool, modules *contracts.Contracts, apiClients *clients.Clients) *Module {
	return &Module{
		pgxDb:   pgxDb,
		modules: modules,
		clients: apiClients,
	}
}

func (m *Module) Name() module.Name {
	return "orders"
}

//go:embed adapters/db/migrations/*.sql
var embedMigrations embed.FS

func (m *Module) Init(ctx context.Context) error {
	restaurantRepo := db.NewRestaurantRepository(m.pgxDb)
	ordersRepo := db.NewOrdersRepository(m.pgxDb)
	customerRepo := db.NewCustomerRepository(m.pgxDb)
	courierRepo := db.NewCourierRepository(m.pgxDb)

	paymentsClient := payments.NewClient(m.clients)

	appService := app.NewService(restaurantRepo, customerRepo, ordersRepo, courierRepo, paymentsClient, m.modules)

	readModel := db.NewReadModel(m.pgxDb)

	httpHandler := http2.NewHandler(
		appService,
		readModel,
		restaurantRepo,
	)
	m.httpHandler = httpHandler

	if err := common.MigrateDatabaseUp(
		ctx,
		string(m.Name()),
		m.pgxDb,
		embedMigrations,
		"adapters/db/migrations",
	); err != nil {
		return err
	}

	return nil
}

func (m *Module) RegisterContracts(ctx context.Context, contracts *contracts.Contracts) error {
	contracts.Orders = ordersModule.Orders{}
	return nil
}

func (m *Module) RegisterHttp(ctx context.Context, e common.EchoRouter) error {
	return http2.Register(ctx, e, m.httpHandler)
}
