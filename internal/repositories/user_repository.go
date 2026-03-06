package repositories

import (
	"context"
	"fmt"

	"socket-flow/internal/models"
	"socket-flow/pkg/postgres"

	"github.com/Masterminds/squirrel"
)

type UserRepo struct {
	pgClient postgres.Client
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	ExistUserByPhoneNumber(ctx context.Context, phoneNumber string) (bool, error)
	GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error)
}

func NewUserRepository(client postgres.Client) *UserRepo {
	return &UserRepo{
		pgClient: client,
	}
}

const users = "users"

var columns = []string{"id", "phone_number", "email"}

func (u *UserRepo) CreateUser(ctx context.Context, user *models.User) error {
	query := u.pgClient.QueryBuilder.Insert(users).
		Columns("phone_number", "password").
		Values(user.PhoneNumber, user.Password)

	sql, args, err := u.pgClient.ToSQL(query)

	if err != nil {
		return fmt.Errorf("failed to build query %w", err)
	}

	if _, err := u.pgClient.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("failed to execute query %w", err)
	}

	return nil
}

func (u *UserRepo) ExistUserByPhoneNumber(ctx context.Context, phoneNumber string) (bool, error) {
	query := u.pgClient.QueryBuilder.
		Select("EXISTS (SELECT 1)").
		From(users).
		Where(squirrel.Eq{"phone_number": phoneNumber})

	sql, args, err := u.pgClient.ToSQL(query)

	if err != nil {
		return false, fmt.Errorf("failed to build query %w", err)
	}

	var exists bool
	if err = u.pgClient.QueryRow(ctx, sql, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to execute query %w", err)
	}

	return exists, nil
}

func (u *UserRepo) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error) {
	query := u.pgClient.QueryBuilder.Select(columns...).
		From(users).
		Where(squirrel.Eq{"phone_number": phoneNumber})

	sql, args, err := u.pgClient.ToSQL(query)

	if err != nil {
		return nil, fmt.Errorf("failed to build query %w", err)
	}

	var user models.User
	if err := u.pgClient.QueryRow(ctx, sql, args...).Scan(&user.Id,
		&user.PhoneNumber,
		&user.Email); err != nil {
		return nil, fmt.Errorf("failed to execute query %w", err)
	}

	return &user, nil
}
