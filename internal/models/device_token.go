package models

import (
	"time"

	"github.com/google/uuid"
)

type DeviceToken struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id" json:"userId"`
	Token     string    `db:"token" json:"token"`
	Platform  string    `db:"platform" json:"platform"` // "android" | "ios" | "web"
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

type RegisterDeviceTokenRequest struct {
	UserID   uuid.UUID `json:"userId"   binding:"required"`
	Token    string    `json:"token"    binding:"required"`
	Platform string    `json:"platform" binding:"required"` // "android" | "ios" | "web"
}
