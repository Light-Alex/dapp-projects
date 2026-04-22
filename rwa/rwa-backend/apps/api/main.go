package main

import (
	"context"
	"os"
	"time"

	"github.com/AnchoredLabs/rwa-backend/apps/api/config"
	"github.com/AnchoredLabs/rwa-backend/apps/api/controller"
	"github.com/AnchoredLabs/rwa-backend/apps/api/server"
	"github.com/AnchoredLabs/rwa-backend/apps/api/service"
	"github.com/AnchoredLabs/rwa-backend/apps/api/types"
	"github.com/AnchoredLabs/rwa-backend/libs/core/evm_helper"
	"github.com/AnchoredLabs/rwa-backend/libs/core/redis_cache"
	"github.com/AnchoredLabs/rwa-backend/libs/core/trade"
	"github.com/AnchoredLabs/rwa-backend/libs/core/web"
	"github.com/AnchoredLabs/rwa-backend/libs/database"
	"github.com/AnchoredLabs/rwa-backend/libs/errors"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	app := &cli.Command{
		Name:        "rwa-api",
		Description: "rwa api",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    types.APP,
				Value:   types.AppApi,
				Aliases: []string{"a"},
				Usage:   "choice the start app: api",
			},
			&cli.StringFlag{
				Name:    types.ConfigFile,
				Value:   types.ConfigFileDefault,
				Aliases: []string{"c"},
				Usage:   "choice the config file path",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliArgs := types.NewCli(c)
			mainCtx := context.WithValue(ctx, log.TraceID, cliArgs.App)
			switch cliArgs.App {
			case types.AppApi:
				return startApp(mainCtx, cliArgs)
			default:
				err := errors.NewErr("invalid app: %s", c.String(types.APP))
				return err.Underlying()
			}
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.ErrorZ(context.Background(), "app run failed", zap.Any("error", err))
	}
}

func startApp(ctx context.Context, cli *types.Cli) error {
	conf, err := config.NewConfig(cli.ConfigFilePath)
	if err != nil {
		log.ErrorZ(ctx, "failed to load config file", zap.String("path", cli.ConfigFilePath), zap.Error(err))
		return err
	}
	log.InitLogger(conf.Logger)
	log.InfoZ(ctx, "load config file success", zap.String("path", cli.ConfigFilePath))

	// Initialize custom validator
	_, err = web.BindGinCusValidator(ctx, "en")
	if err != nil {
		log.ErrorZ(ctx, "failed to initialize custom validator", zap.Error(err))
		return err
	}
	cusLog := fx.WithLogger(func() fxevent.Logger {
		return &fxevent.ZapLogger{Logger: log.Logger()}
	})
	app := fx.New(
		cusLog,
		config.LoadModule(conf),
		server.LoadModule(),
		service.LoadModule(),
		controller.LoadModule(),
		redis_cache.LoadModule(conf.Redis),
		evm_helper.LoadModule(conf.RpcInfo),
		database.LoadModule(conf.Db),
		trade.LoadModule(conf.Alpaca),
		fx.StartTimeout(120*time.Second),
		fx.StopTimeout(120*time.Second),
	)
	app.Run()
	return nil
}
