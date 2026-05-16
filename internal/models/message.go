package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type (
	Message struct {
		ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
		Msg         string        `bson:"msg" json:"msg"`
		From        string        `bson:"from" json:"from"`
		To          string        `bson:"to" json:"to"`
		IsDelivered bool          `bson:"isDelivered" json:"isDelivered"`
		Modified    bool          `bson:"modified" json:"modified"`
		CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
	}
	RequestMessage struct {
		Sender      string `json:"sender"`
		From        string `json:"from"`
		To          string `json:"to" binding:"required"`
		Msg         string `json:"msg"`
		IsDelivered bool   `json:"isDelivered"`
	}

	FindMessagesRequest struct {
		CurrentUserID string `json:"-" form:"-"`
		From          string `json:"from" binding:"required"`
		To            string `json:"to" binding:"required"`
		Limit         uint   `json:"limit" form:"limit,default=20"`
		Offset        uint   `json:"offset" form:"offset,default=0"`
	}
)
