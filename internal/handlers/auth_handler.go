package handlers

import (
	"net/http"
	"socket-flow/internal/errors"
	"socket-flow/internal/models"
	"socket-flow/internal/services"
	"socket-flow/pkg/jwt"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service services.AuthService
}

func NewAuthHandler(service services.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) HandleRegister(c *gin.Context) {
	ctx := c.Request.Context()
	var req models.RegisterUser

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		errors.WriteValidationError(c, err.Error())
		return
	}

	accessToken, refreshToken, phoneNumber, err := h.service.RegisterUser(ctx, &req)
	if err != nil {
		_ = c.Error(err)
		errors.WriteError(c, 0, err)
		return
	}

	if err := jwt.SetAuthCookies(
		c.Writer, accessToken, refreshToken,
	); err != nil {
		_ = c.Error(err)
		errors.WriteError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(
		http.StatusCreated, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"phone_number":  phoneNumber,
		},
	)
}

func (h *AuthHandler) HandleLogin(c *gin.Context) {
	ctx := c.Request.Context()
	var req models.LoginUser

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		errors.WriteValidationError(c, err.Error())
		return
	}

	accessToken, refreshToken, phoneNumber, err := h.service.LoginUser(ctx, req.PhoneNumber, req.Password)
	if err != nil {
		_ = c.Error(err)
		errors.WriteError(c, 0, err)
		return
	}

	if err := jwt.SetAuthCookies(
		c.Writer, accessToken, refreshToken,
	); err != nil {
		_ = c.Error(err)
		errors.WriteError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(
		http.StatusOK, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"phone_number":  phoneNumber,
		},
	)
}

func (h *AuthHandler) HandleLogout(ctx *gin.Context) {
	accessToken, err := ctx.Cookie("access_token")
	if err != nil {
		_ = ctx.Error(err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		_ = ctx.Error(err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.LogoutUser(accessToken, refreshToken); err != nil {
		_ = ctx.Error(err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if err := jwt.DeleteAuthCookies(ctx.Writer); err != nil {
		_ = ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "logout successfull"})
}

func (h *AuthHandler) HandleRefreshToken(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		_ = ctx.Error(err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	accessToken, newRefreshToken, err := h.service.RefreshTokens(refreshToken)
	if err != nil {
		_ = ctx.Error(err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if err := jwt.SetAuthCookies(
		ctx.Writer, accessToken, newRefreshToken,
	); err != nil {
		_ = ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(
		http.StatusCreated, gin.H{
			"access_token":  accessToken,
			"refresh_token": newRefreshToken,
		},
	)
}
