package models

import (
	"github.com/golang-jwt/jwt/v5"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Email     string `gorm:"uniqueIndex;size:255;not null"`
	Password  string `gorm:"size:255;not null"` // note: store bcrypt hash
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"index" json:"user_id"`
	Title     string         `gorm:"size:255" json:"title"`
	SourceURL string         `gorm:"size:1024" json:"source_url"` // original share link
	VideoURL  string         `gorm:"size:1024" json:"video_url"`  // resolved direct video URL or storage path (TODO)
	Status    TaskStatus     `gorm:"size:32;index" json:"status"`
	Progress  string         `json:"progress"`
	FilePath  string         `gorm:"size:1024" json:"file_path"` // local storage path when downloaded
	ErrorMsg  string         `gorm:"size:1024" json:"error_msg"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}
