package repositories

import (
	"context"
	"log/slog"
	"socket-flow/internal/postgres"

	"socket-flow/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	ExistUserByPhoneNumber(ctx context.Context, phoneNumber string) (bool, error)
	GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error)
}

type UserRepositoryImpl struct {
	pgClient     postgres.Client
	QueryBuilder sq.StatementBuilderType
}

const users = "users"

var columns = []string{"id", "phone_number", "email"}

func NewUserRepository(client *postgres.PgClient) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		pgClient:     client,
		QueryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (u *UserRepositoryImpl) CreateUser(ctx context.Context, user *models.User) error {
	sql, args, err := u.QueryBuilder.Insert(users).
		Columns("phone_number", "password").
		Values(user.PhoneNumber, user.Password).
		ToSql()

	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}

	slog.DebugContext(ctx, "query: ", "query", sql, "args:", "args", args)

	_, err = u.pgClient.Exec(ctx, sql, args...)
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

	return nil
}

func (u *UserRepositoryImpl) ExistUserByPhoneNumber(ctx context.Context, phoneNumber string) (bool, error) {
	sql, args, err := u.QueryBuilder.
		Select("EXISTS (SELECT 1)").
		From(users).
		Where(sq.Eq{"phone_number": phoneNumber}).
		ToSql()

	if err != nil {
		return false, errors.Wrap(err, "failed to build query")
	}

	slog.DebugContext(ctx, "query: ", "query", sql, "args:", "args", args)

	var exists bool

	err = u.pgClient.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "failed to execute query")
	}

	return exists, nil
}

func (u *UserRepositoryImpl) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error) {
	sql, args, err := u.QueryBuilder.Select(columns...).
		From(users).
		Where(sq.Eq{"phone_number": phoneNumber}).
		ToSql()

	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}

	slog.DebugContext(ctx, "query: ", "query", sql, "args:", "args", args)

	var user models.User

	err = u.pgClient.QueryRow(ctx, sql, args...).Scan(&user.Id, &user.PhoneNumber, &user.Email)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}

	return &user, nil
}
