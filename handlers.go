package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Simple claims
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// request/response structs
type RegisterReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateTaskReq struct {
	Title string `json:"title" binding:"required"`
	Url   string `json:"url" binding:"required"`
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetPrivacy(c *gin.Context) {
	// TODO: read from DB or file
	c.JSON(http.StatusOK, gin.H{
		"title":   "隐私协议（示例）",
		"content": "这是占位隐私协议。请替换为真实文本或从后端配置加载。",
	})
}

func (h *Handler) GetTerms(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"title":   "服务条款（示例）",
		"content": "这是占位服务条款。请替换为真实文本或从后端配置加载。",
	})
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.svc.CreateUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, _ := generateJWT(u.ID, h.svc.cfg.JWTSecret)
	c.JSON(http.StatusOK, gin.H{"user": gin.H{"id": u.ID, "email": u.Email}, "token": token})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.svc.Authenticate(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	token, _ := generateJWT(u.ID, h.svc.cfg.JWTSecret)
	c.JSON(http.StatusOK, gin.H{"user": gin.H{"id": u.ID, "email": u.Email}, "token": token})
}

func (h *Handler) GetProfile(c *gin.Context) {
	uid := c.GetUint("user_id")
	u, err := h.svc.repo.GetUserByEmail("") // intentionally not used - instead fetch by ID
	// better approach: direct DB query by ID
	_ = u
	_ = err.Error()
	// TODO: implement GetUserByID in repo if needed
	c.JSON(http.StatusOK, gin.H{"id": uid, "email": "placeholder@example.com"})
}

func (h *Handler) CreateTask(c *gin.Context) {
	uid := c.GetUint("user_id")
	var req CreateTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := h.svc.CreateTaskForUser(uid, req.Title, req.Url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": t})
}

func (h *Handler) ListTasks(c *gin.Context) {
	uid := c.GetUint("user_id")
	tasks, err := h.svc.ListTasks(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func (h *Handler) GetTask(c *gin.Context) {
	uid := c.GetUint("user_id")
	id, _ := strconv.Atoi(c.Param("id"))
	t, err := h.svc.GetTask(uid, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": t})
}

func (h *Handler) MarkTaskComplete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var body struct {
		FilePath string `json:"file_path"`
		VideoURL string `json:"video_url"`
	}
	_ = c.BindJSON(&body)
	if err := h.svc.MarkComplete(uint(id), body.FilePath, body.VideoURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) StartTask(c *gin.Context) {
	uid := c.GetUint("user_id")
	id, _ := strconv.Atoi(c.Param("id"))
	t, err := h.svc.GetTask(uid, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	// Kick off background simulated download
	if err := h.svc.StartDownloadTask(context.Background(), t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) StreamTask(c *gin.Context) {
	uid := c.GetUint("user_id")
	id, _ := strconv.Atoi(c.Param("id"))
	t, err := h.svc.GetTask(uid, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	// If task has VideoURL or FilePath, return streaming info.
	// TODO: you may implement proxy streaming, signed URLs, or range requests here.
	c.JSON(http.StatusOK, gin.H{
		"video_url": t.VideoURL,
		"file_path": t.FilePath,
		"status":    t.Status,
		"progress":  t.Progress,
	})
}

// Helper: JWT generation
func generateJWT(userID uint, secret string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// AuthMiddleware extracts JWT and sets user_id in context
func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}
		// expect: Bearer <token>
		var tokenStr string
		if len(auth) > 7 && auth[:7] == "Bearer " {
			tokenStr = auth[7:]
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header"})
			return
		}
		token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		if claims, ok := token.Claims.(*Claims); ok {
			c.Set("user_id", claims.UserID)
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}
		c.Next()
	}
}
