package repositories

import (
	"context"
	"fmt"

	"socket-flow/internal/models"
	"socket-flow/pkg/postgres"

	"github.com/Masterminds/squirrel"
)

type userRepository struct {
	pgClient postgres.Client
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	ExistUserByPhoneNumber(ctx context.Context, phoneNumber string) (bool, error)
	GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error)
}

func NewUserRepository(client postgres.Client) UserRepository {
	return &userRepository{
		pgClient: client,
	}
}

const users = "users"

func (u *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	sql, args, err := u.pgClient.QueryBuilder.Insert(users).
		Columns("phone_number", "password").
		Values(user.PhoneNumber, user.Password).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build query %w", err)
	}

	if _, err := u.pgClient.Exec(sql, args...); err != nil {
		return fmt.Errorf("failed to execute query %w", err)
	}

	return nil
}

func (u *userRepository) ExistUserByPhoneNumber(ctx context.Context, phoneNumber string) (bool, error) {
	sql, args, err := u.pgClient.QueryBuilder.Select("EXISTS (SELECT 1 FROM users WHERE phone_number = ?)").
		ToSql()

	if err != nil {
		return false, fmt.Errorf("failed to build query %w", err)
	}

	var exists bool
	if err = u.pgClient.QueryRow(sql, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to execute query %w", err)
	}
	return exists, nil
}

func (u *userRepository) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error) {
	sql, args, err := u.pgClient.QueryBuilder.Select("*").
		From(users).
		Where(squirrel.Eq{"phone_number": phoneNumber}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query %w", err)
	}

	var user models.User
	if err := u.pgClient.QueryRow(sql, args...).Scan(&user); err != nil {
		return nil, fmt.Errorf("failed to execute query %w", err)
	}
	return &user, nil
}
