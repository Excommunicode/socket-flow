package services

import (
	"context"

	"socket-flow/internal/repositories"

	"github.com/google/uuid"
)

type DeviceTokenService interface {
	Register(ctx context.Context, userID uuid.UUID, token, platform string) error
}

type deviceTokenService struct {
	repo repositories.DeviceTokenRepository
}

func NewDeviceTokenService(repo repositories.DeviceTokenRepository) DeviceTokenService {
	return &deviceTokenService{repo: repo}
}

func (s *deviceTokenService) Register(ctx context.Context, userID uuid.UUID, token, platform string) error {
	return s.repo.Upsert(ctx, userID, token, platform)
}
