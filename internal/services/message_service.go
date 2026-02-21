package services

import (
	"context"
	"fmt"
	"log/slog"
	"socket-flow/internal/models"
	"socket-flow/internal/repositories"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type messageService struct {
	MessageRepository repositories.MessageRepository
}

type MessageService interface {
	CreateMessage(ctx context.Context, msg models.RequestMessage) error
	FindMessages(ctx context.Context, filter models.FindMessagesRequest) ([]models.Message, error)
}

func NewMessageService(messageRepository repositories.MessageRepository) MessageService {
	return &messageService{
		MessageRepository: messageRepository,
	}
}

func (m *messageService) CreateMessage(ctx context.Context, msg models.RequestMessage) error {
	doc := models.Message{
		Msg:         msg.Msg,
		From:        msg.From,
		To:          msg.To,
		IsDelivered: msg.IsDelivered,
		Modified:    false,
		CreatedAt:   time.Now(),
	}

	if err := m.MessageRepository.SaveMessage(ctx, doc); err != nil {
		slog.ErrorContext(ctx, "cannot save the message", "err", err)
		return err
	}
	return nil
}

func (m *messageService) FindMessages(ctx context.Context, filter models.FindMessagesRequest) ([]models.Message, error) {

	from := filter.From.String()
	to := filter.To.String()

	fmt.Println("messages ", from, to)

	messages, err := m.MessageRepository.FindMessages(ctx, filter)
	if err != nil {
		slog.ErrorContext(ctx, "cannot found messages", "err", err)
		return nil, err
	}
	if len(messages) == 0 {
		slog.DebugContext(ctx, "found result is empty")
		return nil, nil
	}

	messageIds := make([]bson.ObjectID, 0, len(messages))
	for _, message := range messages {
		if !message.IsDelivered {
			messageIds = append(messageIds, message.ID)
		}
	}

	if err := m.MessageRepository.MarkAsDeliveredMessages(ctx, messageIds); err != nil {
		slog.ErrorContext(ctx, "cannot mark delivered messages", "err", err)
		return nil, err
	}

	return messages, nil
}
