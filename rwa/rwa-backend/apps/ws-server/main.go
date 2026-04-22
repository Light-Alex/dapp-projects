package main

import (
	"context"
	"os"
	"time"

	"github.com/AnchoredLabs/rwa-backend/apps/ws-server/config"
	"github.com/AnchoredLabs/rwa-backend/apps/ws-server/service"
	"github.com/AnchoredLabs/rwa-backend/apps/ws-server/types"
	"github.com/AnchoredLabs/rwa-backend/apps/ws-server/ws"
	"github.com/AnchoredLabs/rwa-backend/libs/core/kafka_help"
	"github.com/AnchoredLabs/rwa-backend/libs/core/redis_cache"
	"github.com/AnchoredLabs/rwa-backend/libs/errors"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	app := &cli.Command{
		Name:        "ws-server",
		Description: "WebSocket server for RWA backend",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    types.APP,
				Value:   types.AppWs,
				Aliases: []string{"a"},
				Usage:   "choice the start app: ws",
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
			case types.AppWs:
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
	cusLog := fx.WithLogger(func() fxevent.Logger {
		return &fxevent.ZapLogger{Logger: log.Logger()}
	})
	app := fx.New(
		cusLog,
		config.LoadModule(conf),
		ws.LoadModule(),
		service.LoadModule(),
		redis_cache.LoadModule(conf.Redis),
		kafka_help.LoadConsumerModule(conf.Kafka),
		fx.StartTimeout(120*time.Second),
		fx.StopTimeout(120*time.Second),
	)
	app.Run()
	return nil
}


