package models

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type (
	Message struct {
		ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
		Msg         string        `bson:"msg" json:"msg"`
		From        uuid.UUID     `bson:"from" json:"from"` // id
		To          uuid.UUID     `bson:"to" json:"to"`     // id
		IsDelivered bool          `bson:"isDelivered" json:"isDelivered"`
		Modified    bool          `bson:"modified" json:"modified"`
		CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
	}
	RequestMessage struct {
		Sender      string    `json:"sender"`
		From        uuid.UUID `json:"from"`
		To          uuid.UUID `json:"to"` // id
		Msg         string    `json:"msg"`
		IsDelivered bool      `json:"isDelivered"`
	}

	FindMessagesRequest struct {
		From   uuid.UUID `form:"from" binding:"required"`
		To     uuid.UUID `form:"to" binding:"required"`
		Limit  uint      `form:"limit,default=20"`
		Offset uint      `form:"offset,default=0"`
	}
)
