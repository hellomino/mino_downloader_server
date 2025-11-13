package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
	cfg  *Config
	db   *gorm.DB
	rdb  *redis.Client
}

func NewService(db *gorm.DB, rdb *redis.Client, cfg *Config) *Service {
	return &Service{repo: NewRepository(db, rdb), cfg: cfg, db: db, rdb: rdb}
}

// Auth
func (s *Service) CreateUser(email, password string) (*User, error) {
	// validate omitted
	if _, err := s.repo.GetUserByEmail(email); err == nil {
		return nil, errors.New("email exists")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u := &User{Email: email, Password: string(hash)}
	if err := s.repo.CreateUser(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Authenticate(email, password string) (*User, error) {
	u, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}

// Tasks
func (s *Service) CreateTaskForUser(userID uint, title, sourceURL string) (*Task, error) {
	t := &Task{
		UserID:    userID,
		Title:     title,
		SourceURL: sourceURL,
		Status:    StatusPending,
		Progress:  0,
	}
	if err := s.repo.CreateTask(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) ListTasks(userID uint) ([]Task, error) {
	return s.repo.ListTasksByUser(userID)
}

func (s *Service) GetTask(userID uint, id uint) (*Task, error) {
	t, err := s.repo.GetTaskByID(id)
	if err != nil {
		return nil, err
	}
	if t.UserID != userID {
		return nil, errors.New("forbidden")
	}
	return t, nil
}

func (s *Service) StartDownloadTask(ctx context.Context, t *Task) error {
	// NOTE: Real download/parse logic should be implemented by user.
	// Here we simulate background progress with a goroutine and update DB.
	if t.Status == StatusRunning {
		return errors.New("already running")
	}
	t.Status = StatusRunning
	_ = s.repo.UpdateTask(t)

	go func(taskID uint) {
		// Simulate work
		for i := 1; i <= 10; i++ {
			time.Sleep(1 * time.Second)
			tt, err := s.repo.GetTaskByID(taskID)
			if err != nil {
				return
			}
			tt.Progress = float32(i) / 10.0
			_ = s.repo.UpdateTask(tt)
		}
		// After simulated download, mark completed and set a fake file path / video url
		tt, _ := s.repo.GetTaskByID(taskID)
		tt.Status = StatusCompleted
		tt.Progress = 1.0
		tt.FilePath = fmt.Sprintf("/data/videos/%d.mp4", tt.ID)                   // TODO: change to real storage path
		tt.VideoURL = fmt.Sprintf("https://cdn.example.com/videos/%d.mp4", tt.ID) // TODO: actual CDN or signed URL
		_ = s.repo.UpdateTask(tt)
	}(t.ID)

	return nil
}

func (s *Service) MarkComplete(taskID uint, filePath, videoURL string) error {
	t, err := s.repo.GetTaskByID(taskID)
	if err != nil {
		return err
	}
	t.Status = StatusCompleted
	t.Progress = 1.0
	t.FilePath = filePath
	t.VideoURL = videoURL
	return s.repo.UpdateTask(t)
}
