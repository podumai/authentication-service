package service

import (
	"authentication_service/internal/database"
	"authentication_service/internal/logger"
	"context"
	"database/sql"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

const (
	insertTokenQuery = "INSERT INTO bs.tokens (user_id, refresh_token) VALUES ($1, $2)"
	removeTokenQuery = "DELETE FROM bs.tokens WHERE user_id = $1 AND refresh_token = $2"
)

type TokenService interface {
	Insert(ctx context.Context, key string, token string) error
	Remove(ctx context.Context, key string, token string) error
}

type tokenService struct {
	db     *gorm.DB
	logger logger.Logger
}

func (ts *tokenService) Insert(ctx context.Context, key string, token string) error {
	tr := otel.Tracer("TokenService")
	ctx, span := tr.Start(ctx, "Insert")
	span.SetAttributes(attribute.String("Key", key))
	defer span.End()

	sqlDB, err := ts.db.DB()
	if err != nil {
		return err
	}

	trx, err := sqlDB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer safeRollback(trx, ts.logger)

	_, err = trx.ExecContext(ctx, insertTokenQuery, key, token)
	if err != nil {
		return err
	}
	err = trx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (ts *tokenService) Remove(ctx context.Context, key string, token string) error {
	tr := otel.Tracer("TokenService")
	ctx, span := tr.Start(ctx, "Retrieve")
	span.SetAttributes(attribute.String("Key", key))
	defer span.End()

	sqlDB, err := ts.db.DB()
	if err != nil {
		return err
	}

	trx, err := sqlDB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer safeRollback(trx, ts.logger)

	result, err := trx.ExecContext(ctx, removeTokenQuery, key, token)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return err
	} else if affected != 1 {
		return errors.New("illegal refresh token")
	}

	err = trx.Commit()
	if err != nil {
		return err
	}

	return nil
}

type TokenOpts struct {
	Database database.DatabaseService
	Logger   logger.Logger
}

func NewTokenService(opts *TokenOpts) TokenService {
	return &tokenService{
		db:     opts.Database.DB(),
		logger: opts.Logger,
	}
}
