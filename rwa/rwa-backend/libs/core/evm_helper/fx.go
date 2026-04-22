package evm_helper

import "go.uber.org/fx"

func LoadModule(infoMap RpcInfoMap) fx.Option {
	return fx.Module("evmChain", fx.Supply(infoMap), fx.Provide(
		NewEvmClient,
	))
}
