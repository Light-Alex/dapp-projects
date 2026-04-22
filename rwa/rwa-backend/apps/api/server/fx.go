package server

import "go.uber.org/fx"

func LoadModule() fx.Option {
	return fx.Module("server", fx.Provide(
		NewRouter,
		NewServer,
	),
		fx.Invoke(func(_ *Server) {}),
	)
}
