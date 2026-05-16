package middlewares

import (
	"socket-flow/internal/auth"
	errs "socket-flow/internal/errors"

	"github.com/gin-gonic/gin"
)

func KeycloakMiddleware(authenticator *auth.KeycloakAuthenticator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString, ok := auth.BearerToken(ctx.GetHeader("Authorization"))
		if !ok {
			errs.WriteError(ctx, 0, errs.ErrAuthHeaderMissing)
			ctx.Abort()
			return
		}

		user, err := authenticator.Validate(ctx.Request.Context(), tokenString)
		if err != nil {
			errs.WriteError(ctx, 0, errs.ErrUnauthorizedToken)
			ctx.Abort()
			return
		}

		auth.SetUser(ctx, user)
		ctx.Next()
	}
}
