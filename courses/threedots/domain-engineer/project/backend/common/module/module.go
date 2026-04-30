package module

import (
	"context"

	"eats/backend/common"
	"eats/backend/common/module/contracts"
)

type Name string

type Module interface {
	Name() Name
	Init(ctx context.Context) error
	RegisterHttp(ctx context.Context, e common.EchoRouter) error
	RegisterContracts(ctx context.Context, contracts *contracts.Contracts) error
}
