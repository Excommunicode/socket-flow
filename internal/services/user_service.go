package services

import (
	"context"
	"log/slog"
	"socket-flow/internal/models"
	"socket-flow/internal/postgres"
	"socket-flow/internal/repositories"

	"github.com/pkg/errors"
)

type UserServiceImpl struct {
	repo        repositories.UserRepository
	transaction postgres.Transactor
}

type UserService interface {
	GetUserByPhone(ctx context.Context, email string) (*models.UserResponse, error)
}

func NewUserService(transaction postgres.Transactor, repo repositories.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{
		transaction: transaction,
		repo:        repo,
	}
}

func (u UserServiceImpl) GetUserByPhone(ctx context.Context, phone string) (*models.UserResponse, error) {

	result := new(models.User)

	err := u.transaction.WithinROTransaction(ctx, func(ctx context.Context) error {
		var err error
		result, err = u.repo.GetUserByPhoneNumber(ctx, models.NormalizePhone(phone))

		return err
	})

	if err != nil {
		slog.ErrorContext(ctx, "failed to get user by phone", "phone", phone, "error", err)

		return nil, errors.Wrap(err, "get user by phone")
	}

	return &models.UserResponse{
		Id:          result.Id,
		Email:       result.Email,
		PhoneNumber: &result.PhoneNumber,
		Role:        result.Role,
	}, nil

}
