package service

import "go.uber.org/fx"

func LoadModule() fx.Option {
	return fx.Module("service",
		fx.Provide(
			NewOrderSyncService,
			NewAlpacaWebSocketService,
		),
		fx.Invoke(startAlpacaWebSocketService),
	)
}

// startAlpacaWebSocketService triggers the lifecycle hooks of AlpacaWebSocketService
// Without this invoke, the OnStart/OnStop hooks registered in NewAlpacaWebSocketService
// would never execute because fx only runs lifecycle hooks for invoked types.
func startAlpacaWebSocketService(wsService *AlpacaWebSocketService) {
	// The lifecycle hooks are registered in NewAlpacaWebSocketService via lc.Append()
	// By requesting the wsService as a parameter, fx will execute those hooks
}

