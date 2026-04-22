package service

import (
	"go.uber.org/fx"
)

func LoadModule() fx.Option {
	return fx.Module("service",
		fx.Provide(
			NewOrderUpdateSubscriber,
			NewBarUpdateSubscriber,
		),
		fx.Invoke(
			func(_ *OrderUpdateSubscriber) {},
			func(_ *BarUpdateSubscriber) {},
		),
	)
}
