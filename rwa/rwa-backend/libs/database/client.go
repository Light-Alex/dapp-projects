package database

import (
	"context"
	"fmt"
	"time"

	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/acmestack/gorm-plus/gplus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DbConf struct {
	Host                     string          `json:"host" mapstructure:"host" yaml:"host"`
	Port                     int             `json:"port" mapstructure:"port" yaml:"port"`
	Username                 string          `json:"username" mapstructure:"username" yaml:"username"`
	Password                 string          `json:"password" mapstructure:"password" yaml:"password"`
	Database                 string          `mapstructure:"database" yaml:"database"`
	MaxIdleConns             int             `mapstructure:"maxIdleConns" mapstructure_default:"5" yaml:"maxIdleConns"`
	MaxOpenConns             int             `mapstructure:"maxOpenConns" mapstructure_default:"50" yaml:"maxOpenConns"`
	LogLevel                 logger.LogLevel `mapstructure:"logLevel" mapstructure_default:"2" yaml:"logLevel"`
	DefaultStringSize        uint            `mapstructure:"defaultStringSize" mapstructure_default:"255" yaml:"defaultStringSize"`
	ConnMaxLifetime          int             `mapstructure:"connMaxLifetime" mapstructure_default:"3600" yaml:"connMaxLifetime"`
	SQLSlowThresholdMill     int             `mapstructure:"sqlSlowThresholdMill" mapstructure_default:"200" yaml:"SQLSlowThresholdMill"`
	AutoMigrate              bool            `mapstructure:"autoMigrate" yaml:"autoMigrate"`
	DisableNestedTransaction bool            `mapstructure:"disableNestedTransaction" yaml:"disableNestedTransaction"`
	MigrationPath            string          `mapstructure:"migrationPath" yaml:"migrationPath"`
	MigrationToVersion       uint            `mapstructure:"migrationToVersion" yaml:"migrationToVersion"`
	SslMode                  string          `mapstructure:"sslMode" yaml:"sslMode"`
}

// FillDefault fill default
func (d *DbConf) FillDefault() {
	if d.Host == "" || d.Username == "" || d.Password == "" || d.Database == "" {
		panic("connect database params that host or user or password or database may be one of the invalid")
	}
	if d.MaxIdleConns == 0 {
		d.MaxIdleConns = 5
	}
	if d.MaxOpenConns == 0 {
		d.MaxOpenConns = 50
	}
	if d.LogLevel == 0 {
		d.LogLevel = logger.Error
	}
	if d.DefaultStringSize == 0 {
		d.DefaultStringSize = 255
	}
	if d.ConnMaxLifetime == 0 {
		d.ConnMaxLifetime = 3600 // one hour
	}
	if d.SQLSlowThresholdMill == 0 {
		d.SQLSlowThresholdMill = 2000 // 200 millisecond
	}
	if d.SslMode == "" {
		d.SslMode = "enable"
	}
}

// Dsn dsn
func (d *DbConf) Dsn() string {
	//nolint:lll
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s", d.Host, d.Username, d.Password, d.Database, d.Port, d.SslMode)
}

// NewPgDb new pg db
func NewPgDb(dbConf *DbConf) (*gorm.DB, error) {
	dbConf.FillDefault()
	newLogger := log.NewDbLogger(
		logger.Config{
			SlowThreshold:             time.Millisecond * time.Duration(dbConf.SQLSlowThresholdMill),
			LogLevel:                  dbConf.LogLevel,
			IgnoreRecordNotFoundError: true,
		})
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dbConf.Dsn(),
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		AllowGlobalUpdate:                        false,
		DisableNestedTransaction:                 dbConf.DisableNestedTransaction,
		Logger:                                   newLogger,
		PrepareStmt:                              true,
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(dbConf.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConf.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(dbConf.ConnMaxLifetime))
	return db, nil
}

// BindGormPlus bind gorm plus
func BindGormPlus(db *gorm.DB) {
	gplus.Init(db)
}

// InitModel init model
func InitModel(ctx context.Context, db *gorm.DB, autoMigrate bool, t ...interface{}) error {
	if !autoMigrate {
		return nil
	}
	if err := db.AutoMigrate(t...); err != nil {
		log.ErrorZ(ctx, "Db AutoMigrate failed")
		return err
	}
	return nil
}
