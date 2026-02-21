package handlers

import (
	"net/http"
	"socket-flow/internal/errors"
	"socket-flow/internal/models"
	"socket-flow/internal/services"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	MessageService services.MessageService
}

func NewMessageHandler(messageService services.MessageService) *MessageHandler {
	return &MessageHandler{
		MessageService: messageService,
	}
}

func (m *MessageHandler) FindMessage(c *gin.Context) {
	ctx := c.Request.Context()
	body := models.FindMessagesRequest{}

	if err := c.ShouldBindJSON(&body); err != nil {
		_ = c.Error(err)
		errors.WriteValidationError(c, err.Error())
	}

	result, err := m.MessageService.FindMessages(ctx, body)
	if err != nil {
		_ = c.Error(err)
		errors.WriteError(c, 0, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
