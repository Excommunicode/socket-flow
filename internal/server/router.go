package server

import (
	"socket-flow/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func initRouters(handler *Handler) *gin.Engine {
	r := gin.Default()

	r.Use(middlewares.LoggerMiddleware())
	r.Use(middlewares.CORSMiddleware())

	r.GET("/ws", gin.WrapF(handler.SocketHandler.ServeMs))

	api := r.Group("/api")
	api.Use(middlewares.KeycloakMiddleware(handler.Authenticator))

	messageRouter := api.Group("/messages")
	{
		messageRouter.POST("", handler.MessageHandler.FindMessage)
	}

	userRouter := api.Group("/users")
	{
		userRouter.GET("/me", handler.UserHandler.GetCurrentUser)
	}

	deviceTokenRouter := api.Group("/device-tokens")
	{
		deviceTokenRouter.POST("", handler.DeviceTokenHandler.RegisterDeviceToken)
	}

	return r
}
