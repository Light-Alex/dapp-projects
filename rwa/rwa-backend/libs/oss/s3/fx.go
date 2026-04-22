package s3

import (
	"go.uber.org/fx"
)

func LoadModule(conf *Config) fx.Option {
	return fx.Module("s3", fx.Supply(conf),
		fx.Provide(
			NewService,
		),
	)
}
