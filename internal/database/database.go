package database

import (
	"authentication_service/internal/config"
	"authentication_service/internal/logger"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Opts struct {
	Config *config.Database
	Logger logger.Logger
}

type DatabaseService interface {
	DB() *gorm.DB
	Close() error
}

type Database struct {
	db     *gorm.DB
	Logger logger.Logger
}

func (d *Database) DB() *gorm.DB {
	return d.db
}

func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func NewDatabase(opts *Opts) (DatabaseService, error) {
	var (
		db  *gorm.DB
		err error
	)

	switch opts.Config.Driver {
	case "postgres":
		db, err = gorm.Open(postgres.Open(opts.Config.DSN), &gorm.Config{})
	case "sqlite":
		if opts.Config.DSN == "" {
			return nil, fmt.Errorf("invalid DSN: sqlite requires a non empty DSN")
		}
		db, err = gorm.Open(sqlite.Open(opts.Config.DSN), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported database driver %s", opts.Config.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %v", opts.Config.Driver, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get db instance: %v", err)
	}

	sqlDB.SetMaxIdleConns(opts.Config.PoolMaxIdleConns)
	sqlDB.SetMaxOpenConns(opts.Config.PoolMaxOpenConns)
	sqlDB.SetConnMaxLifetime(opts.Config.PoolConnMaxLifetime)
	opts.Logger.Info("Database connected", logger.Field{Key: "driver", Value: opts.Config.Driver})

	return &Database{db: db, Logger: opts.Logger}, nil
}
