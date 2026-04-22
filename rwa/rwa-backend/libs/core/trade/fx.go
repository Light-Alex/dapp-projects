package trade

import (
	"go.uber.org/fx"
)

func LoadModule(config *AlpacaConfig) fx.Option {
	opts := []fx.Option{
		fx.Module("trade"),
	}
	if config != nil {
		opts = append(opts,
			fx.Supply(config),
			fx.Provide(
				NewAlpacaService,
				func(svc *AlpacaService) TradeService {
					return svc
				},
			),
		)
	} else {
		// Provide nil TradeService when config is not available
		opts = append(opts,
			fx.Provide(func() TradeService {
				return nil
			}),
		)
	}
	return fx.Options(opts...)
}
