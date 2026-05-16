package services

import (
	"context"

	"socket-flow/internal/repositories"
)

type DeviceTokenService interface {
	Register(ctx context.Context, userID string, token, platform string) error
}

type deviceTokenService struct {
	repo repositories.DeviceTokenRepository
}

func NewDeviceTokenService(repo repositories.DeviceTokenRepository) DeviceTokenService {
	return &deviceTokenService{repo: repo}
}

func (s *deviceTokenService) Register(ctx context.Context, userID string, token, platform string) error {
	return s.repo.Upsert(ctx, userID, token, platform)
}
