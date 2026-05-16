package handlers

import (
	"net/http"
	"socket-flow/internal/auth"
	"socket-flow/internal/errors"
	"socket-flow/internal/models"
	"socket-flow/internal/services"

	"github.com/gin-gonic/gin"
)

type DeviceTokenHandler struct {
	DeviceTokenService services.DeviceTokenService
}

func NewDeviceTokenHandler(deviceTokenService services.DeviceTokenService) *DeviceTokenHandler {
	return &DeviceTokenHandler{DeviceTokenService: deviceTokenService}
}

func (h *DeviceTokenHandler) RegisterDeviceToken(c *gin.Context) {
	ctx := c.Request.Context()

	body := new(models.RegisterDeviceTokenRequest)

	err := c.ShouldBindJSON(body)

	if err != nil {
		_ = c.Error(err)
		errors.WriteValidationError(c, err.Error())
		return
	}

	user, err := auth.UserFromContext(c)
	if err != nil {
		_ = c.Error(err)
		errors.WriteError(c, 0, err)
		return
	}

	err = h.DeviceTokenService.Register(ctx, user.Subject, body.Token, body.Platform)

	if err != nil {
		_ = c.Error(err)
		errors.WriteError(c, 0, err)
		return
	}

	c.Status(http.StatusNoContent)
}
