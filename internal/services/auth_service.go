package services

import (
	"context"
	"database/sql"
	stdErrors "errors"
	"log/slog"
	"socket-flow/internal/postgres"

	errs "socket-flow/internal/errors"
	"socket-flow/internal/jwt"
	"socket-flow/internal/models"
	"socket-flow/internal/repositories"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type authService struct {
	repo            repositories.UserRepository
	transaction     postgres.Transactor
	tokenRepository repositories.TokenRepository
}

type AuthService interface {
	RegisterUser(ctx context.Context, user *models.RegisterUser) (accessToken, refreshToken, phoneNumber string, err error)
	LoginUser(ctx context.Context, phoneNumber, password string) (accessToken, refreshToken, phone string, err error)
	LogoutUser(accessToken, refreshToken string) error
	RefreshTokens(refreshToken string) (string, string, error)
	ValidateJwtToken(userId uuid.UUID, tokenString string) (bool, error)
}

func NewAuthService(transaction postgres.Transactor,
	repo repositories.UserRepository,
	tokenRepository repositories.TokenRepository) AuthService {
	return &authService{
		transaction:     transaction,
		repo:            repo,
		tokenRepository: tokenRepository,
	}
}

func (s *authService) RegisterUser(ctx context.Context, req *models.RegisterUser) (string, string, string, error) {
	log := slog.With("phone_number", req.PhoneNumber, "op", "RegisterUser")

	log.InfoContext(ctx, "attempting to register new user")

	req.PhoneNumber = models.NormalizePhone(req.PhoneNumber)

	var existingUser bool
	if err := s.transaction.WithinRWTransaction(ctx, func(ctx context.Context) error {
		var err error
		existingUser, err = s.repo.ExistUserByPhoneNumber(ctx, req.PhoneNumber)
		return err
	}); err != nil {
		log.ErrorContext(ctx, "database check failed", "err", err)
		return "", "", "", errors.Wrap(err, "check existing user")
	}

	if existingUser {
		log.WarnContext(ctx, "registration failed: user already exists")

		return "", "", "", errs.ErrUserExists
	}

	hashedPassword, err := jwt.HashPassword(req.Password)
	if err != nil {
		log.ErrorContext(ctx, "password hashing failed", "err", err)

		return "", "", "", errors.Wrap(err, "hash password")
	}

	user := &models.User{
		PhoneNumber: req.PhoneNumber,
		Password:    hashedPassword,
	}

	if err := s.transaction.WithinRWTransaction(ctx, func(ctx context.Context) error {
		if err := s.repo.CreateUser(ctx, user); err != nil {
			return errors.Wrap(err, "create user")
		}
		return nil
	}); err != nil {
		log.ErrorContext(ctx, "failed to save user to database", "err", err)
		return "", "", "", errors.Wrap(err, "save user")
	}

	access, refresh, err := s.generateAndStoreTokens(user.Id, user.Role)
	if err != nil {
		log.ErrorContext(ctx, "token generation failed", "err", err, "user_id", user.Id)
		return "", "", "", errors.Wrap(err, "generate and store tokens")
	}

	log.InfoContext(ctx, "user registered successfully", "user_id", user.Id)
	return access, refresh, user.PhoneNumber, nil
}

func (s *authService) LoginUser(ctx context.Context, phoneNumber, password string) (string, string, string, error) {
	log := slog.With("phone_number", phoneNumber, "op", "LoginUser")

	log.DebugContext(ctx, "login attempt started")

	phoneNumber = models.NormalizePhone(phoneNumber)

	user, err := s.repo.GetUserByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			log.WarnContext(ctx, "login failed: user not found")
			return "", "", "", errs.ErrUserNotFound
		}
		log.ErrorContext(ctx, "repository error during login", "err", err)
		return "", "", "", errs.ErrUserVerifyFailed
	}

	checkPassword, err := jwt.CheckPassword(password, user.Password)
	if err != nil {
		log.ErrorContext(ctx, "internal password check error", "err", err)
		return "", "", "", errs.ErrInvalidCreds
	}

	if !checkPassword {
		log.WarnContext(ctx, "login failed: invalid password", "user_id", user.Id)
		return "", "", "", errs.ErrInvalidCreds
	}

	access, refresh, err := s.generateAndStoreTokens(user.Id, user.Role)
	if err != nil {
		log.ErrorContext(ctx, "token generation failed", "err", err, "user_id", user.Id)
		return "", "", "", errors.Wrap(err, "generate and store tokens")
	}

	log.InfoContext(ctx, "user logged in successfully", "user_id", user.Id)
	return access, refresh, user.PhoneNumber, nil
}

func (s *authService) LogoutUser(accessToken, refreshToken string) error {
	claims, err := jwt.ParseJWT(accessToken)
	if err != nil {
		return errs.ErrTokenParsingFailed
	}

	userID, convErr := uuid.Parse(claims.Subject)
	if convErr != nil {
		return errs.ErrInvalidTokenSub
	}

	if err := s.tokenRepository.DeleteJWTokens(
		userID, accessToken, refreshToken,
	); err != nil {
		return errs.ErrTokenDeletionFailed
	}

	return nil
}

func (s *authService) RefreshTokens(refreshToken string) (
	string, string, error,
) {
	claims, err := jwt.ParseJWT(refreshToken)
	if err != nil {
		return "", "", errs.ErrTokenParsingFailed
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return "", "", errs.ErrInvalidTokenSub
	}
	if err := s.tokenRepository.DeleteJWToken(userID, refreshToken); err != nil {
		return "", "", errs.ErrTokenDeletionFailed
	}

	accessToken, newRefreshToken, err := s.generateAndStoreTokens(
		userID, models.ParseRole(claims.Role),
	)
	if err != nil {
		return "", "", errors.Wrap(err, "generate and store tokens")
	}

	if err := s.tokenRepository.SaveJWTokens(
		userID, accessToken, newRefreshToken,
	); err != nil {
		return "", "", errs.ErrTokenStorage
	}

	return accessToken, newRefreshToken, nil
}

func (s *authService) ValidateJwtToken(userId uuid.UUID, tokenString string) (bool, error) {
	isValid, err := s.tokenRepository.ValidateJWToken(userId, tokenString)
	if err != nil {
		return false, nil
	}

	return isValid, nil
}

func (s *authService) generateAndStoreTokens(userID uuid.UUID, role models.Role) (
	string, string, error,
) {
	uidStr := userID
	accessToken, err := jwt.GenerateJWT(uidStr.String(), 15, role.String())
	if err != nil {
		return "", "", errs.ErrTokenGeneration
	}

	refreshToken, err := jwt.GenerateJWT(uidStr.String(), 1440, role.String())
	if err != nil {
		return "", "", errs.ErrTokenGeneration
	}

	if err := s.tokenRepository.SaveJWTokens(
		userID, accessToken, refreshToken,
	); err != nil {
		return "", "", errs.ErrTokenStorage
	}

	return accessToken, refreshToken, nil
}
