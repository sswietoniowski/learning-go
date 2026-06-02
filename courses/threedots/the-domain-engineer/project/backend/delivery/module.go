package delivery

import (
	"context"

	"eats/backend/common"
	"eats/backend/common/module"
	"eats/backend/common/module/contracts"
	deliveryModule "eats/backend/delivery/api/module"
	"eats/backend/delivery/app"
)

type Module struct {
	service *app.Service
}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() module.Name {
	return "delivery"
}

func (m *Module) Init(ctx context.Context) error {
	m.service = app.NewService()
	return nil
}

func (m *Module) RegisterContracts(ctx context.Context, contracts *contracts.Contracts) error {
	contracts.Delivery = deliveryModule.New(m.service)
	return nil
}

func (m *Module) RegisterHttp(ctx context.Context, e common.EchoRouter) error {
	// this module doesn't expose any HTTP endpoints
	return nil
}
