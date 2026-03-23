package repositories

import (
	"context"
	"log/slog"
	"socket-flow/internal/models"
	"socket-flow/internal/postgres"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type DeviceTokenRepository interface {
	Upsert(ctx context.Context, userID uuid.UUID, token, platform string) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.DeviceToken, error)
	DeleteByToken(ctx context.Context, token string) error
}

type DeviceTokenRepositoryImpl struct {
	pgClient     postgres.Client
	queryBuilder sq.StatementBuilderType
}

func NewDeviceTokenRepository(client postgres.Client) *DeviceTokenRepositoryImpl {
	return &DeviceTokenRepositoryImpl{
		pgClient:     client,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

const deviceTokensTable = "device_tokens"

func (r *DeviceTokenRepositoryImpl) Upsert(ctx context.Context, userID uuid.UUID, token, platform string) error {
	now := time.Now()

	const query = `
		INSERT INTO device_tokens (user_id, token, platform, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (token) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			platform = EXCLUDED.platform,
			updated_at = EXCLUDED.updated_at`

	_, err := r.pgClient.Exec(ctx, query, userID, token, platform, now, now)

	if err != nil {
		return errors.Wrap(err, "upsert device token")
	}

	return nil
}

func (r *DeviceTokenRepositoryImpl) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.DeviceToken, error) {
	sql, args, err := r.queryBuilder.
		Select("id", "user_id", "token", "platform", "created_at", "updated_at").
		From(deviceTokensTable).
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "build find query")
	}

	slog.DebugContext(ctx, "query", "sql", sql, "args", args)

	var tokens []models.DeviceToken

	err = r.pgClient.Select(ctx, &tokens, sql, args...)

	if err != nil {
		return nil, errors.Wrap(err, "find device tokens by user_id")
	}

	return tokens, nil
}

func (r *DeviceTokenRepositoryImpl) DeleteByToken(ctx context.Context, token string) error {
	sql, args, err := r.queryBuilder.
		Delete(deviceTokensTable).
		Where(sq.Eq{"token": token}).
		ToSql()

	if err != nil {
		return errors.Wrap(err, "build delete query")
	}

	_, err = r.pgClient.Exec(ctx, sql, args...)

	if err != nil {
		return errors.Wrap(err, "delete device token")
	}

	return nil
}
