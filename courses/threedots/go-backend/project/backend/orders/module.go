package orders

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/common"
	"eats/backend/common/module"
	"eats/backend/common/module/contracts"
	http2 "eats/backend/orders/api/http"
	ordersModule "eats/backend/orders/api/module"
)

type Module struct {
	pgxDb       *pgxpool.Pool
	httpHandler http2.Handler

	modules *contracts.Contracts
}

func NewModule(pgxDb *pgxpool.Pool, modules *contracts.Contracts) *Module {
	return &Module{
		pgxDb:   pgxDb,
		modules: modules,
	}
}

func (m *Module) Name() module.Name {
	return "orders"
}

func (m *Module) Init(ctx context.Context) error {
	httpHandler := http2.NewHandler()
	m.httpHandler = httpHandler

	return nil
}

func (m *Module) RegisterContracts(ctx context.Context, contracts *contracts.Contracts) error {
	contracts.Orders = ordersModule.Orders{}
	return nil
}

func (m *Module) RegisterHttp(ctx context.Context, e common.EchoRouter) error {
	return http2.Register(ctx, e, m.httpHandler)
}
