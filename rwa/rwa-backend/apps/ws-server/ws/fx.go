package ws

import (
	"go.uber.org/fx"
)

func LoadModule() fx.Option {
	return fx.Module("ws",
		fx.Provide(
			NewServer,
			NewSubUnsubService,
		),
		fx.Invoke(func(_ *Server, _ *SubUnsubService) {}),
	)
}

