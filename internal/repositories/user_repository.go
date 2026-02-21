package repositories

import (
	"context"
	"database/sql"
	"errors"
	"socket-flow/internal/db"
	"socket-flow/internal/models"
	"socket-flow/pkg/postgres"
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

func (u *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	queries := db.QueriesFromContext(ctx, u.pgClient.Db)

	err := queries.CreateUser(ctx, db.CreateUserParams{
		PhoneNumber: user.PhoneNumber,
		Password:    user.Password,
	})
	return err
}

func (u *userRepository) ExistUserByPhoneNumber(ctx context.Context, phoneNumber string) (bool, error) {
	queries := db.QueriesFromContext(ctx, u.pgClient.Db)

	exists, err := queries.ExistUserByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (u *userRepository) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error) {
	queries := db.QueriesFromContext(ctx, u.pgClient.Db)

	dbUser, err := queries.GetUserByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}

	return &models.User{
		Id:          dbUser.ID,
		Username:    &dbUser.Username.String,
		Email:       &dbUser.Email.String,
		PhoneNumber: dbUser.PhoneNumber,
		Role:        models.Role(dbUser.Role),
		Password:    dbUser.Password,
		CreatedAt:   dbUser.CreatedAt,
		UpdateAt:    &dbUser.UpdatedAt.Time,
	}, nil
}
