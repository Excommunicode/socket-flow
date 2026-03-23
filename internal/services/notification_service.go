package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"socket-flow/internal/repositories"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const fcmSendURL = "https://fcm.googleapis.com/v1/projects/%s/messages:send"

type NotificationService interface {
	SendPushNotification(ctx context.Context, toUserID uuid.UUID, title, body string) error
}

type notificationService struct {
	deviceTokenRepo repositories.DeviceTokenRepository
	fcmProjectID    string
	fcmAccessToken  string // OAuth2 Bearer token (см. примечание ниже)
	httpClient      *http.Client
}

func NewNotificationService(
	deviceTokenRepo repositories.DeviceTokenRepository,
	fcmProjectID string,
	fcmAccessToken string,
) NotificationService {
	return &notificationService{
		deviceTokenRepo: deviceTokenRepo,
		fcmProjectID:    fcmProjectID,
		fcmAccessToken:  fcmAccessToken,
		httpClient:      &http.Client{},
	}
}

type fcmMessage struct {
	Message fcmPayload `json:"message"`
}

type fcmPayload struct {
	Token        string            `json:"token"`
	Notification fcmNotification   `json:"notification"`
	Data         map[string]string `json:"data,omitempty"`
}

type fcmNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (s *notificationService) SendPushNotification(
	ctx context.Context,
	toUserID uuid.UUID,
	title, body string,
) error {
	if s.fcmProjectID == "" || s.fcmAccessToken == "" {
		slog.DebugContext(ctx, "FCM not configured, skipping push", "userId", toUserID)
		return nil
	}

	tokens, err := s.deviceTokenRepo.FindByUserID(ctx, toUserID)
	if err != nil {
		return errors.Wrap(err, "find device tokens")
	}

	if len(tokens) == 0 {
		slog.DebugContext(ctx, "no device tokens for user", "userId", toUserID)
		return nil
	}

	url := fmt.Sprintf(fcmSendURL, s.fcmProjectID)

	for _, dt := range tokens {
		if err := s.sendToToken(ctx, url, dt.Token, title, body); err != nil {
			// Если токен протух — удаляем
			if isInvalidTokenError(err) {
				slog.WarnContext(ctx, "removing invalid FCM token", "token", dt.Token)
				_ = s.deviceTokenRepo.DeleteByToken(ctx, dt.Token)
			} else {
				slog.ErrorContext(ctx, "fcm send error", "err", err, "token", dt.Token)
			}
		}
	}

	return nil
}

func (s *notificationService) sendToToken(ctx context.Context, url, token, title, body string) error {
	payload := fcmMessage{
		Message: fcmPayload{
			Token: token,
			Notification: fcmNotification{
				Title: title,
				Body:  body,
			},
			Data: map[string]string{
				"type": "new_message",
			},
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "marshal fcm payload")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return errors.Wrap(err, "build fcm request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.fcmAccessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "send fcm request")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("fcm unauthorized")
	}
	if resp.StatusCode == http.StatusNotFound {
		return errInvalidToken
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("fcm unexpected status: %d", resp.StatusCode)
	}

	return nil
}

var errInvalidToken = errors.New("invalid fcm token")

func isInvalidTokenError(err error) bool {
	return err == errInvalidToken
}
