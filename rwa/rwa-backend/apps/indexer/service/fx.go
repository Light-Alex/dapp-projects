package service

import (
	"context"
	"crypto/ecdsa"

	"github.com/AnchoredLabs/rwa-backend/apps/indexer/config"
	"github.com/AnchoredLabs/rwa-backend/apps/indexer/service/handlers"
	"github.com/AnchoredLabs/rwa-backend/libs/core/evm_helper"
	"github.com/AnchoredLabs/rwa-backend/libs/core/trade"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/ethereum/go-ethereum/crypto"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func LoadModule() fx.Option {
	return fx.Module("service",
		fx.Provide(
			NewBlockService,
			NewEventListener,
			NewProcessTxService,
			provideHandlers,
		),
		fx.Invoke(registerEventListener),
	)
}

func provideHandlers(
	tradeService trade.TradeService,
	conf *config.Config,
	evmClient *evm_helper.EvmClient,
) ([]EventHandler, error) {
	// Parse backend private key for on-chain transactions (e.g., cancelOrder on Alpaca failure)
	var backendPK *ecdsa.PrivateKey
	if conf.Backend != nil && conf.Backend.PrivateKey != "" {
		pk, err := crypto.HexToECDSA(conf.Backend.PrivateKey)
		if err != nil {
			log.ErrorZ(context.Background(), "failed to parse backend private key", zap.Error(err))
		} else {
			backendPK = pk
		}
	}

	orderSubmittedHandler, err := handlers.NewHandleOrderSubmitted(tradeService, conf.Chain.UsdmAddress, evmClient, conf.Chain.ChainId, conf.Chain.PocAddress, backendPK)
	if err != nil {
		return nil, err
	}

	orderCancelledHandler, err := handlers.NewHandleOrderCancelled(tradeService)
	if err != nil {
		return nil, err
	}

	cancelRequestedHandler, err := handlers.NewHandleCancelRequested(tradeService)
	if err != nil {
		return nil, err
	}

	orderExecutedHandler, err := handlers.NewHandleOrderExecuted(tradeService)
	if err != nil {
		return nil, err
	}

	// TODO: Add Gate contract event handlers once Gate Go bindings are generated.
	// Gate.sol emits these events: PendingDeposit, DepositProcessed, PendingWithdraw,
	// WithdrawalProcessed, MinimumDepositAmountSet, MinimumWithdrawalAmountSet,
	// DepositsPaused, DepositsUnpaused, WithdrawalsPaused, WithdrawalsUnpaused.
	// Steps:
	// 1. Generate Gate ABI JSON from the contract
	// 2. Run: abigen -abi Gate.abi.json -out gate.go -pkg rwa -type Gate
	// 3. Create handlers for PendingDeposit, DepositProcessed, PendingWithdraw, WithdrawalProcessed
	// 4. Register handlers here with ContractTypeGate

	return []EventHandler{
		orderSubmittedHandler,
		orderCancelledHandler,
		cancelRequestedHandler,
		orderExecutedHandler,
	}, nil
}

func registerEventListener(
	lc fx.Lifecycle,
	eventListener *EventListener,
	blockService *BlockService,
	processService *ProcessTxService,
	conf *config.Config,
) {
	// Create an independent context for polling that outlives the OnStart ctx.
	pollingCtx, pollingCancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Initialize process service cache
			if err := processService.InitCache(ctx); err != nil {
				log.ErrorZ(ctx, "failed to init process service cache", zap.Error(err))
				pollingCancel()
				return err
			}

			// Start polling for events using the independent context
			go func() {
				if err := eventListener.StartPolling(pollingCtx, blockService, processService); err != nil {
					log.ErrorZ(pollingCtx, "event listener polling stopped with error", zap.Error(err))
				}
			}()

			log.InfoZ(ctx, "event listener started",
				zap.Uint64("chainId", conf.Chain.ChainId),
				zap.String("pocAddress", conf.Chain.PocAddress))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.InfoZ(ctx, "stopping event listener")
			pollingCancel()
			return nil
		},
	})
}
