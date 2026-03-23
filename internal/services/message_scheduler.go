package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"socket-flow/internal/repositories"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

type MessageScheduler struct {
	ttl               time.Duration
	cleanupCron       string
	MessageRepository repositories.MessageRepository
	cronObject        *cron.Cron
}

func NewMessageScheduler(c *cron.Cron, ttl string, cleanupCron string, messageRepository repositories.MessageRepository) (
	*MessageScheduler, error) {

	_, err := cron.ParseStandard(cleanupCron)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("invalid cron expression: %v", cleanupCron))
	}

	parsedTTL, err := parseTTL(ttl)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("invalid ttl: %v", ttl))
	}

	return &MessageScheduler{
		cronObject:        c,
		ttl:               parsedTTL,
		cleanupCron:       cleanupCron,
		MessageRepository: messageRepository,
	}, nil
}

func (m *MessageScheduler) StartCleanupScheduler(ctx context.Context) error {

	_, err := m.cronObject.AddFunc(m.cleanupCron, m.CleanMessages)
	if err != nil {
		return errors.Wrap(err, "failed to add cleanup cron job")
	}

	m.cronObject.Start()

	<-ctx.Done()
	m.cronObject.Stop()

	return nil
}

func (m *MessageScheduler) CleanMessages() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	threshold := time.Now().Add(-m.ttl)

	err := m.MessageRepository.DeleteMessageByCreatedAt(ctx, threshold)

	if err != nil {
		slog.Info(fmt.Sprintf("cannot delete messages err: %s", err))
	}
}

func parseTTL(ttl string) (time.Duration, error) {
	if ttl == "" {
		return 0, errors.New("ttl cannot be empty")
	}

	d, err := time.ParseDuration(ttl)
	if err != nil {
		return 0, err
	}

	return d, nil

}
