package router

import (
	"github.com/gin-gonic/gin"
	"minodl/controller"
	"minodl/ws"
	"net/http"
)

func ProxyApi() *gin.Engine {
	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	// public
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	router.GET("/policy/privacy", controller.GetPrivacy)
	router.GET("/policy/terms", controller.GetTerms)
	router.GET("/pay/plan", nil)
	router.GET("/connect/to", ws.Connect)

	router.POST("/pay/notify_callback", nil)

	return router
}
