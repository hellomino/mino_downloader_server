package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env if present
	_ = godotenv.Load()

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := InitGorm(cfg)
	if err != nil {
		log.Fatalf("db init: %v", err)
	}

	rdb := InitRedis(cfg)

	// Auto-migrate (simple)
	if err := AutoMigrate(db); err != nil {
		log.Fatalf("auto-migrate: %v", err)
	}

	svc := NewService(db, rdb, cfg)
	h := NewHandler(svc)

	router := gin.Default()

	// public
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	router.GET("/policy/privacy", h.GetPrivacy)
	router.GET("/policy/terms", h.GetTerms)
	router.POST("/auth/register", h.Register)
	router.POST("/auth/login", h.Login)

	// protected
	auth := router.Group("/", AuthMiddleware(cfg.JWTSecret))
	{
		auth.GET("/me", h.GetProfile)
		auth.POST("/tasks", h.CreateTask)
		auth.GET("/tasks", h.ListTasks)
		auth.GET("/tasks/:id", h.GetTask)
		auth.POST("/tasks/:id/complete", h.MarkTaskComplete) // used when server-side download finishes
		auth.POST("/tasks/:id/start", h.StartTask)           // kick off mock download job
		auth.GET("/tasks/:id/stream", h.StreamTask)          // returns video URL or proxy (TODO)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		log.Printf("listening %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
