package database

import (
	"context"

	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func LoadModule(conf *DbConf) fx.Option {
	return fx.Module("database", fx.Supply(conf),
		fx.Provide(
			NewPgDb,
		),
		fx.Invoke(func(db *gorm.DB, dbConf *DbConf) {
			ctx := context.Background()
			BindGormPlus(db)
			migration, err := NewMigration(ctx, dbConf)
			if err != nil {
				log.ErrorZ(ctx, "NewMigration failed", zap.Error(err))
				panic(err)
			}
			err = migration.Up(ctx)
			if err != nil {
				log.ErrorZ(ctx, "migration up failed", zap.Error(err))
				panic(err)
			}
		}),
	)
}
