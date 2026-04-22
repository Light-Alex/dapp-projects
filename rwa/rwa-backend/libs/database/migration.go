package database

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

type Migration struct {
	client *migrate.Migrate
}

// NewMigration todo log should ignore password
func NewMigration(ctx context.Context, conf *DbConf) (*Migration, error) {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s&x-migrations-table=%s",
		conf.Username,
		conf.Password,
		net.JoinHostPort(conf.Host, strconv.Itoa(conf.Port)),
		conf.Database,
		conf.SslMode,
		"_migrations",
	)
	client, err := migrate.New(conf.MigrationPath, dbURL)
	if err != nil {
		log.ErrorZ(ctx, "new migration failed", zap.Error(err))
		return nil, err
	}
	return &Migration{
		client: client,
	}, nil
}

func (m *Migration) To(ctx context.Context, targetVersion uint) error {
	if targetVersion == 0 {
		log.WarnZ(ctx, "target version is 0")
		return nil
	}
	if err := m.client.Migrate(targetVersion); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.ErrorZ(ctx, "Migration to version failed", zap.Error(err), zap.Uint("version", targetVersion))
		return err
	}
	afterVersion, _, _ := m.client.Version()
	log.InfoZ(ctx, "Migration to version success", zap.Uint("version", afterVersion))
	return nil
}

func (m *Migration) Up(ctx context.Context) error {
	if err := m.client.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.ErrorZ(ctx, "Migration up version failed", zap.Error(err))
		return err
	}
	afterVersion, _, _ := m.client.Version()
	log.InfoZ(ctx, "Migration up version success", zap.Uint("version", afterVersion))
	return nil
}

func (m *Migration) Down(ctx context.Context) error {
	if err := m.client.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.ErrorZ(ctx, "Migration down version failed", zap.Error(err))
		return err
	}
	version, _, _ := m.client.Version()
	log.InfoZ(ctx, "Migration down version success", zap.Uint("version", version))
	return nil
}
