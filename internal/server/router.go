package server

import (
	"socket-flow/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func initRouters(handler *Handler) *gin.Engine {
	router := gin.Default()

	router.Use(middlewares.LoggerMiddleware())
	router.Use(middlewares.CORSMiddleware())

	router.GET("/ws", gin.WrapF(handler.SocketHandler.ServeMs))

	api := router.Group("/api")
	authRoutes := api.Group("/auth")

	authHandler := handler.AuthHandler

	{
		authRoutes.POST("/register", authHandler.HandleRegister)
		authRoutes.POST("/login", authHandler.HandleLogin)
		authRoutes.POST("/logout", authHandler.HandleLogout)
		authRoutes.POST("/refresh", authHandler.HandleRefreshToken)
	}

	messageRouter := api.Group("/messages")

	//messageRouter.Use(middlewares.JWTMiddleware())
	{
		messageRouter.POST("", handler.MessageHandler.FindMessage)
	}

	userRouter := api.Group("/users")
	//userRouter.Use(middlewares.JWTMiddleware())
	{
		userRouter.POST("", handler.UserHandler.GetUserByPhone)
	}

	return router
}
