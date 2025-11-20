package service

import (
	"context"
	"errors"
	"fmt"
	"minodl/dao"
	"minodl/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Auth
func CreateUser(email, password string) (*models.User, error) {
	// validate omitted
	if _, err := dao.GetUserByEmail(email); err == nil {
		return nil, errors.New("email exists")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u := &models.User{Email: email, Password: string(hash)}
	if err := dao.CreateUser(u); err != nil {
		return nil, err
	}
	return u, nil
}

func Authenticate(email, password string) (*models.User, error) {
	u, err := dao.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}

// Tasks
func CreateTaskForUser(userID uint, sourceURL string) (*models.Task, error) {
	videoInfo, err := ParseVideoInfo(sourceURL)
	if err != nil {
		return nil, err
	}
	t := &models.Task{
		UserID:    userID,
		Title:     videoInfo.Title,
		SourceURL: sourceURL,
		Status:    models.StatusPending,
		Progress:  "0",
	}
	if err := dao.CreateTask(t); err != nil {
		return nil, err
	}
	return t, nil
}

func ListTasks(userID uint) ([]models.Task, error) {
	return dao.ListTasksByUser(userID)
}

func GetTask(userID uint, id uint) (*models.Task, error) {
	t, err := dao.GetTaskByID(id)
	if err != nil {
		return nil, err
	}
	if t.UserID != userID {
		return nil, errors.New("forbidden")
	}
	return t, nil
}

func StartDownloadTask(ctx context.Context, t *models.Task) error {
	// NOTE: Real download/parse logic should be implemented by user.
	// Here we simulate background progress with a goroutine and update DB.
	if t.Status == models.StatusRunning {
		return errors.New("already running")
	}
	t.Status = models.StatusRunning
	_ = dao.UpdateTask(t)

	go func(taskID uint) {
		// Simulate work
		for i := 1; i <= 10; i++ {
			time.Sleep(1 * time.Second)
			tt, err := dao.GetTaskByID(taskID)
			if err != nil {
				return
			}
			tt.Progress = fmt.Sprintf("%.2f", float32(i)/10.0)
			_ = dao.UpdateTask(tt)
		}
		// After simulated download, mark completed and set a fake file path / video url
		tt, _ := dao.GetTaskByID(taskID)
		tt.Status = models.StatusCompleted
		tt.Progress = "100"
		tt.FilePath = fmt.Sprintf("/data/videos/%d.mp4", tt.ID)                   // TODO: change to real storage path
		tt.VideoURL = fmt.Sprintf("https://cdn.example.com/videos/%d.mp4", tt.ID) // TODO: actual CDN or signed URL
		_ = dao.UpdateTask(tt)
	}(t.ID)

	return nil
}

func MarkComplete(taskID uint, filePath, videoURL string) error {
	t, err := dao.GetTaskByID(taskID)
	if err != nil {
		return err
	}
	t.Status = models.StatusCompleted
	t.Progress = "100"
	t.FilePath = filePath
	t.VideoURL = videoURL
	return dao.UpdateTask(t)
}
