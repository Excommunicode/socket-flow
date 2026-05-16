package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"socket-flow/internal/auth"
	"socket-flow/internal/models"

	"github.com/gin-gonic/gin"
)

type recordingMessageService struct {
	filter models.FindMessagesRequest
}

func (s *recordingMessageService) CreateMessage(ctx context.Context, msg models.RequestMessage) error {
	return nil
}

func (s *recordingMessageService) FindMessages(
	ctx context.Context,
	filter models.FindMessagesRequest,
) ([]models.Message, error) {
	s.filter = filter
	return []models.Message{}, nil
}

func TestMessageHandlerFindMessageUsesAuthenticatedSubject(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	messageService := new(recordingMessageService)
	handler := NewMessageHandler(messageService)

	router := gin.New()
	router.POST("/messages", func(c *gin.Context) {
		auth.SetUser(c, auth.AuthenticatedUser{Subject: "keycloak-subject"})
		handler.FindMessage(c)
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/messages",
		bytes.NewBufferString(`{"from":"spoofed-user","to":"recipient","limit":10,"offset":0}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", resp.Code, http.StatusOK)
	}
	if messageService.filter.CurrentUserID != "keycloak-subject" {
		t.Fatalf("CurrentUserID = %q; want keycloak-subject", messageService.filter.CurrentUserID)
	}
	if messageService.filter.From != "keycloak-subject" {
		t.Fatalf("From = %q; want keycloak-subject", messageService.filter.From)
	}
	if messageService.filter.To != "recipient" {
		t.Fatalf("To = %q; want recipient", messageService.filter.To)
	}
}
