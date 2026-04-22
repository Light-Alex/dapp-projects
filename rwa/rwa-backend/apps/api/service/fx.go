package service

import (
	"go.uber.org/fx"
)

func LoadModule() fx.Option {
	return fx.Module("service", fx.Provide(
		NewStockService,
		NewOrderService,
	))
}
