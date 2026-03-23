package handlers

import (
	"net/http"
	"socket-flow/internal/errors"
	"socket-flow/internal/models"
	"socket-flow/internal/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (u *UserHandler) GetUserByPhone(c *gin.Context) {

	ctx := c.Request.Context()

	req := new(models.UserRequest)

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		errors.WriteValidationError(c, err.Error())
		return
	}

	user, err := u.userService.GetUserByPhone(ctx, req.PhoneNumber)
	if err != nil {
		_ = c.Error(err)
		errors.WriteError(c, 0, err)
		return
	}

	c.JSON(http.StatusOK, user)
}
