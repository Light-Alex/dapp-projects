package config

import (
	"go.uber.org/fx"
)

func LoadModule(conf *Config) fx.Option {
	return fx.Module("config",
		fx.Provide(func() *Config {
			return conf
		}),
	)
}

