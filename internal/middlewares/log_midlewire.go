package middlewares

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		duration := time.Since(start)

		message := fmt.Sprintf(
			"Request Log:\n  %-8s: %s\n  %-8s: %s\n  %-8s: %d\n  %-8s: %s",
			"Method", ctx.Request.Method,
			"Path", ctx.Request.URL.Path,
			"Status", ctx.Writer.Status(),
			"Duration", duration.String(),
		)

		if len(ctx.Errors) > 0 {
			for _, e := range ctx.Errors {
				message += fmt.Sprintf("\n  %-8s: %s", "Error", e.Error())
			}
			slog.Error(message)
		} else {
			slog.Info(message)
		}
	}
}
