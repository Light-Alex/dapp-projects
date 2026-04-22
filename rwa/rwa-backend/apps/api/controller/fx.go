package controller

import (
	"go.uber.org/fx"
)

func LoadModule() fx.Option {
	return fx.Module("controller", fx.Provide(
		// Simple controllers - no special dependencies
		NewCommonController,
		NewTradeController,
		NewStockController,
		NewOrderController,
	))
}
