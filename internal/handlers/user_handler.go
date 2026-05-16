package handlers

import (
	"net/http"
	"socket-flow/internal/auth"
	"socket-flow/internal/errors"

	"github.com/gin-gonic/gin"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (u *UserHandler) GetCurrentUser(c *gin.Context) {
	user, err := auth.UserFromContext(c)
	if err != nil {
		_ = c.Error(err)
		errors.WriteError(c, 0, err)
		return
	}

	c.JSON(http.StatusOK, user)
}
