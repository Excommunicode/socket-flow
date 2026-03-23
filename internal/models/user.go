package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type (
	User struct {
		Id          uuid.UUID  `db:"id"`
		Username    *string    `db:"username"`
		Email       *string    `db:"email"`
		PhoneNumber string     `db:"phone_number"`
		Role        Role       `db:"role"`
		Password    string     `db:"password" json:"-"`
		CreatedAt   time.Time  `db:"created_at"`
		UpdateAt    *time.Time `db:"updated_at"`
	}

	UserRequest struct {
		PhoneNumber string `json:"phoneNumber"`
	}

	RegisterUser struct {
		PhoneNumber string `json:"phoneNumber"`
		Password    string `json:"Password"`
	}

	UserResponse struct {
		Id          uuid.UUID `json:"id"`
		Username    string    `json:"username"`
		Email       *string   `json:"email"`
		PhoneNumber *string   `db:"phone_number"`
		Role        Role      `db:"role"`
	}

	LoginUser struct {
		PhoneNumber string `json:"phone_number"`
		Password    string `json:"password"`
	}

	Role int
)

const (
	user Role = iota + 1
	admin
)

func (r Role) String() string {
	switch r {
	case user:
		return "user"
	case admin:
		return "admin"
	default:
		return "unknown"
	}
}

func ParseRole(s string) Role {
	switch s {
	case "user":
		return user
	case "admin":
		return admin
	default:
		return 0
	}
}

func NormalizePhone(phone string) string {
	return strings.Join(strings.Fields(phone), "")
}
