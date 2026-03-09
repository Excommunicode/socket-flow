package middlewares

import (
	"net/http"
	"socket-flow/internal/jwt"
	"strings"

	errs "socket-flow/internal/errors"

	"github.com/gin-gonic/gin"
)

func JWTMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{"error": errs.ErrAuthHeaderMissing.Error()},
			)
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jwt.ParseJWT(tokenString)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}

		ctx.Set("claims", claims)
		ctx.Set("tokenString", tokenString)
		ctx.Next()
	}
}
