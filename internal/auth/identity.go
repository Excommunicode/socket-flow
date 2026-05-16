package auth

import (
	"socket-flow/internal/errors"

	"github.com/gin-gonic/gin"
)

const contextUserKey = "authenticated_user"

type AuthenticatedUser struct {
	Subject  string   `json:"sub"`
	Username string   `json:"username,omitempty"`
	Email    string   `json:"email,omitempty"`
	Roles    []string `json:"roles,omitempty"`
}

func SetUser(c *gin.Context, user AuthenticatedUser) {
	c.Set(contextUserKey, user)
}

func UserFromContext(c *gin.Context) (AuthenticatedUser, error) {
	value, ok := c.Get(contextUserKey)
	if !ok {
		return AuthenticatedUser{}, errors.ErrClaimsNotFound
	}

	user, ok := value.(AuthenticatedUser)
	if !ok || user.Subject == "" {
		return AuthenticatedUser{}, errors.ErrInvalidClaims
	}

	return user, nil
}
