package main

import (
	"context"
	"os"
	"time"

	"github.com/AnchoredLabs/rwa-backend/apps/indexer/config"
	"github.com/AnchoredLabs/rwa-backend/apps/indexer/service"
	"github.com/AnchoredLabs/rwa-backend/apps/indexer/types"
	"github.com/AnchoredLabs/rwa-backend/libs/core/evm_helper"
	"github.com/AnchoredLabs/rwa-backend/libs/core/trade"
	"github.com/AnchoredLabs/rwa-backend/libs/database"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	app := &cli.Command{
		Name:        "rwa-indexer",
		Description: "RWA Indexer service to parse on-chain data directly from RPC",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    types.ConfigFile,
				Value:   types.ConfigFileDefault,
				Aliases: []string{"c"},
				Usage:   "choice the config file path",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliArgs := types.NewCli(c)
			mainCtx := context.WithValue(ctx, log.TraceID, "rwa-indexer")
			return startIndexer(mainCtx, cliArgs)
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.ErrorZ(context.Background(), "app run failed", zap.Any("error", err))
	}
}

func startIndexer(ctx context.Context, cli *types.Cli) error {
	conf, err := config.NewConfig(cli.ConfigFilePath)
	if err != nil {
		log.ErrorZ(ctx, "failed to load config file", zap.String("path", cli.ConfigFilePath), zap.Error(err))
		return err
	}
	log.InitLogger(conf.Logger)
	log.InfoZ(ctx, "load config file success", zap.String("path", cli.ConfigFilePath))
	cusLog := fx.WithLogger(func() fxevent.Logger {
		return &fxevent.ZapLogger{Logger: log.Logger()}
	})
	app := fx.New(
		cusLog,
		config.LoadModule(conf),
		evm_helper.LoadModule(conf.RpcInfo),
		database.LoadModule(conf.Db),
		trade.LoadModule(conf.Alpaca),
		service.LoadModule(),
		fx.StartTimeout(120*time.Second),
		fx.StopTimeout(120*time.Second),
	)
	log.InfoZ(ctx, "Begin to start rwa-indexer-direct service")
	app.Run()
	return nil
}
