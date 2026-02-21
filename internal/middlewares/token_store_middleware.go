package middlewares

import (
	"net/http"
	errs "socket-flow/internal/errors"
	"socket-flow/internal/services"
	"socket-flow/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var _ = jwt.Claims{}

func TokenStoreMiddleware(authService services.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claimsVal, exists := ctx.Get("claims")
		if !exists {
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{"error": errs.ErrClaimsNotFound.Error()},
			)
			ctx.Abort()
			return
		}

		claims, ok := claimsVal.(*jwt.Claims)
		if !ok {
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{"error": errs.ErrInvalidClaims.Error()},
			)
			ctx.Abort()
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{"error": errs.ErrUnauthorizedToken.Error()},
			)
			ctx.Abort()
			return
		}

		tokenStringVal, exists := ctx.Get("tokenString")
		if !exists {
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{"error": errs.ErrTokenNotFound.Error()},
			)
			ctx.Abort()
			return
		}

		tokenString, ok := tokenStringVal.(string)
		if !ok {
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{"error": errs.ErrTokenTypeFailed.Error()},
			)
			ctx.Abort()
			return
		}

		valid, err := authService.ValidateJwtToken(userID, tokenString)
		if err != nil || !valid {
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{"error": errs.ErrUnauthorizedToken.Error()},
			)
			ctx.Abort()
			return
		}

		ctx.Set("claims", claims)
		ctx.Next()
	}
}
