package repositories

import (
	"context"
	"log/slog"
	"socket-flow/internal/config"
	"socket-flow/internal/models"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MessageRepository interface {
	SaveMessage(ctx context.Context, msg models.Message) error
	FindMessages(ctx context.Context, filter models.FindMessagesRequest) ([]models.Message, error)
	MarkAsDeliveredMessages(ctx context.Context, messageIds []bson.ObjectID) error
	DeleteMessages(ctx context.Context, messageIds []bson.ObjectID)
	DeleteMessageByCreatedAt(ctx context.Context, ttl time.Time) error
}

type MessageRepo struct {
	messageCollection *mongo.Collection
}

func NewMessageRepository(client *mongo.Client, cfg config.MongoConfig) *MessageRepo {
	const messageCollection = "message"

	collection := client.Database(cfg.Database).Collection(messageCollection)

	return &MessageRepo{
		messageCollection: collection,
	}
}

func (m *MessageRepo) SaveMessage(ctx context.Context, msg models.Message) error {

	_, err := m.messageCollection.InsertOne(ctx, msg)

	if err != nil {
		return errors.Wrap(err, "cannot insert the message")
	}

	return nil
}

func (m *MessageRepo) FindMessages(ctx context.Context, filter models.FindMessagesRequest) ([]models.Message, error) {
	filters := bson.M{
		"$or": bson.A{
			bson.M{
				"from": filter.From,
				"to":   filter.To,
			},
			bson.M{
				"from": filter.To,
				"to":   filter.From,
			},
		},
	}

	opts := options.Find().
		SetLimit(int64(filter.Limit)).
		SetSkip(int64(filter.Offset)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := m.messageCollection.Find(ctx, filters, opts)
	if err != nil {
		return nil, errors.Wrap(err, "find messages")
	}

	defer func() {
		if err := cursor.Close(ctx); err != nil {
			slog.ErrorContext(ctx, "cannot close the cursor ", "err", err)
		}
	}()

	result := make([]models.Message, 0, cursor.RemainingBatchLength())

	for cursor.Next(ctx) {
		var message models.Message
		if err := cursor.Decode(&message); err != nil {
			return nil, errors.Wrap(err, "decode message")
		}

		result = append(result, message)
	}
	return result, nil
}

func (m *MessageRepo) MarkAsDeliveredMessages(ctx context.Context, messageIds []bson.ObjectID) error {

	filter := bson.M{
		"_id": bson.M{
			"$in": messageIds,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"isDelivered": true,
		},
	}

	_, err := m.messageCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return errors.Wrap(err, "mark messages as delivered")
	}

	return nil
}

func (m *MessageRepo) DeleteMessages(ctx context.Context, messageIds []bson.ObjectID) {

}

func (m *MessageRepo) DeleteMessageByCreatedAt(ctx context.Context, ttl time.Time) error {
	filter := bson.M{
		"created_at": bson.M{
			"$lt": ttl,
		},
	}

	_, err := m.messageCollection.DeleteMany(ctx, filter)
	if err != nil {
		return errors.Wrap(err, "failed to delete messages older")
	}

	return nil
}
