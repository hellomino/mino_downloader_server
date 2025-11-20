package controller

import (
	"minodl/config"
	"minodl/dao"
	"minodl/models"
	"minodl/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

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
	Url string `json:"url" binding:"required"`
}

func GetPrivacy(c *gin.Context) {
	// TODO: read from DB or file
	c.JSON(http.StatusOK, gin.H{
		"title":   "隐私协议（示例）",
		"content": "这是占位隐私协议。请替换为真实文本或从后端配置加载。",
	})
}

func GetTerms(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"title":   "服务条款（示例）",
		"content": "这是占位服务条款。请替换为真实文本或从后端配置加载。",
	})
}

func Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := service.CreateUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, _ := generateJWT(u.ID, config.Get().JWTSecret)
	c.JSON(http.StatusOK, gin.H{"user": gin.H{"id": u.ID, "email": u.Email}, "token": token})
}

func Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := service.Authenticate(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	token, _ := generateJWT(u.ID, config.Get().JWTSecret)
	c.JSON(http.StatusOK, gin.H{"user": gin.H{"id": u.ID, "email": u.Email}, "token": token})
}

func GetProfile(c *gin.Context) {
	uid := c.GetUint("user_id")
	u, err := dao.GetUserById(int64(uid))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid uid"})
		return
	}
	c.JSON(http.StatusOK, u)
}

func CreateTask(c *gin.Context) {
	uid := c.GetUint("user_id")
	var req CreateTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := service.CreateTaskForUser(uid, req.Url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": t})
}

func ListTasks(c *gin.Context) {
	uid := c.GetUint("user_id")
	tasks, err := service.ListTasks(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func GetTask(c *gin.Context) {
	uid := c.GetUint("user_id")
	id, _ := strconv.Atoi(c.Param("id"))
	t, err := service.GetTask(uid, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": t})
}

func MarkTaskComplete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var body struct {
		FilePath string `json:"file_path"`
		VideoURL string `json:"video_url"`
	}
	_ = c.BindJSON(&body)
	if err := service.MarkComplete(uint(id), body.FilePath, body.VideoURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func StreamTask(c *gin.Context) {
	uid := c.GetUint("user_id")
	id, _ := strconv.Atoi(c.Param("id"))
	t, err := service.GetTask(uid, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if t.Status == models.StatusCompleted {
		c.JSON(http.StatusConflict, gin.H{"msg": "this task is already completed"})
		return
	}
	HandleStream(c, t)
}

func StartTask(c *gin.Context) {
	// TODO save to cloud disk
	/*uid := c.GetUint("user_id")
	id, _ := strconv.Atoi(c.Param("id"))
	t, err := service.GetTask(uid, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := service.StartDownloadTask(context.Background(), t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})*/
}

// Helper: JWT generation
func generateJWT(userID uint, secret string) (string, error) {
	claims := models.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
