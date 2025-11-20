package router

import (
	"github.com/gin-gonic/gin"
	"minodl/controller"
	"minodl/middleware"
	"net/http"
)

func MinoAPI() *gin.Engine {
	router := gin.Default()
	// public
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	router.GET("/policy/privacy", controller.GetPrivacy)
	router.GET("/policy/terms", controller.GetTerms)
	router.POST("/auth/register", controller.Register)
	router.POST("/auth/login", controller.Login)

	// protected
	auth := router.Group("/api", middleware.RequestAuthMiddleware())
	{
		auth.GET("/me", controller.GetProfile)
		auth.POST("/tasks", controller.CreateTask)
		auth.GET("/tasks", controller.ListTasks)
		auth.GET("/tasks/:id", controller.GetTask)
		auth.POST("/tasks/:id/complete", controller.MarkTaskComplete) // used when server-side download finishes
		auth.POST("/tasks/:id/start", controller.StartTask)           // kick off mock download job
		auth.GET("/tasks/:id/stream", controller.StreamTask)
	}
	return router
}
