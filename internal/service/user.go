package service

import (
	"authentication_service/internal/cache"
	"authentication_service/internal/database"
	"authentication_service/internal/database/model"
	"authentication_service/internal/logger"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type UserService interface {
	Create(ctx context.Context, email, passwordHash string) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	UpdatePassword(ctx context.Context, id int, passwordHash []byte) error
	Delete(ctx context.Context, id int) error
	Shutdown() error
}

type userService struct {
	db     *gorm.DB
	cache  cache.CacheService
	logger logger.Logger
}

type UserServiceOpts struct {
	Database database.DatabaseService
	Cache    cache.CacheService
	Logger   logger.Logger
}

const (
	createUserQuery      = "INSERT INTO bs.users (role, status, email, password) VALUES ((SELECT id FROM bs.roles WHERE title = 'user'), 'active', $1, $2)"
	findUserByEmailQuery = "SELECT u.id, password, permissions, status, created_at, updated_at FROM bs.users AS u INNER JOIN bs.roles AS r ON u.role = r.id WHERE email = $1"
	updateStatusQuery    = "UPDATE db.users SET status = $1 WHERE id = $2"
	updatePasswordQuery  = "UPDATE db.users SET password = $1 WHERE id = $2"
	deleteUserQuery      = "DELETE db.users WHERE id = $1"
)

func NewUserService(opts *UserServiceOpts) UserService {
	return &userService{
		db:     opts.Database.DB(),
		cache:  opts.Cache,
		logger: opts.Logger,
	}
}

func safeRollback(trx *sql.Tx, l logger.Logger) {
	if err := recover(); err != nil {
		if trxErr := trx.Rollback(); trxErr != nil && errors.Is(trxErr, sql.ErrTxDone) {
			l.Error("rollback after panic", logger.Field{Key: "error", Value: fmt.Sprintf("trx err: %s; (original: %v)", trxErr, err)})
		}
	}
	if err := trx.Rollback(); err != nil && errors.Is(err, sql.ErrTxDone) {
		l.Error("transation failed", logger.Field{Key: "error", Value: err.Error()})
	}
}

func (us *userService) Create(ctx context.Context, email, passwordHash string) error {
	tr := otel.Tracer("UserService")
	ctx, span := tr.Start(ctx, "Create")
	span.SetAttributes(attribute.String("Email", email))
	defer span.End()

	sqlDB, err := us.db.DB()
	if err != nil {
		return err
	}
	trx, err := sqlDB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer safeRollback(trx, us.logger)

	_, err = trx.ExecContext(ctx, createUserQuery, email, passwordHash)
	if err != nil {
		return err
	}
	err = trx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (us *userService) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	tr := otel.Tracer("UserService")
	ctx, span := tr.Start(ctx, "FindByEmail")
	span.SetAttributes(attribute.String("Email", email))
	defer span.End()

	cacheKey := "user:" + email
	user := &model.User{}

	if cached, err := us.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		if err := json.Unmarshal([]byte(cached), &user); err != nil {
			return nil, err
		}
		return user, nil
	}

	sqlDB, err := us.db.DB()
	if err != nil {
		return nil, err
	}

	perms := ""

	err = sqlDB.QueryRowContext(ctx, findUserByEmailQuery, email).Scan(
		&user.ID,
		&user.PasswordHash,
		&perms,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	user.Permissions = strings.Split(perms[1:len(perms)-1], ",")
	data, _ := json.Marshal(user)
	if err := us.cache.Set(ctx, cacheKey, data, time.Minute); err != nil {
		return nil, err
	}
	return user, nil
}

func (us *userService) UpdateStatus(ctx context.Context, id int, status string) error {
	tr := otel.Tracer("UserService")
	ctx, span := tr.Start(ctx, "UpdateStatus")
	span.SetAttributes(attribute.Int("ID", id))
	defer span.End()

	sqlDB, err := us.db.DB()
	if err != nil {
		return err
	}

	trx, err := sqlDB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer safeRollback(trx, us.logger)

	_, err = trx.ExecContext(ctx, updateStatusQuery, status, id)
	if err != nil {
		return err
	}
	err = trx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (us *userService) UpdatePassword(ctx context.Context, id int, passwordHash []byte) error {
	tr := otel.Tracer("UserService")
	ctx, span := tr.Start(ctx, "UpdatePassword")
	span.SetAttributes(attribute.Int("ID", id))
	defer span.End()

	sqlDB, err := us.db.DB()
	if err != nil {
		return err
	}

	trx, err := sqlDB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer safeRollback(trx, us.logger)

	_, err = trx.ExecContext(ctx, updatePasswordQuery, passwordHash, id)
	if err != nil {
		return err
	}
	err = trx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (us *userService) Delete(ctx context.Context, id int) error {
	tr := otel.Tracer("UserService")
	ctx, span := tr.Start(ctx, "Delete")
	span.SetAttributes(attribute.Int("ID", id))
	defer span.End()

	sqlDB, err := us.db.DB()
	if err != nil {
		return err
	}

	trx, err := sqlDB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer safeRollback(trx, us.logger)

	_, err = trx.ExecContext(ctx, deleteUserQuery, id)
	if err != nil {
		return err
	}
	err = trx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (us *userService) Shutdown() error {
	sqlDB, err := us.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
