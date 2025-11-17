package dao

import (
	"context"
	"minodl/mdb"
	"minodl/models"
	"time"
)

func CreateUser(u *models.User) error {
	return mdb.Mysql.Create(u).Error
}

func GetUserByEmail(email string) (*models.User, error) {
	var u models.User
	if err := mdb.Mysql.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func CreateTask(t *models.Task) error {
	return mdb.Mysql.Create(t).Error
}

func GetTaskByID(id uint) (*models.Task, error) {
	var t models.Task
	if err := mdb.Mysql.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func ListTasksByUser(userID uint) ([]models.Task, error) {
	var out []models.Task
	if err := mdb.Mysql.Where("user_id = ?", userID).Order("created_at desc").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func UpdateTask(t *models.Task) error {
	t.UpdatedAt = time.Now()
	return mdb.Mysql.Save(t).Error
}

// Simple cache helper (JSON could be used). Example usage: caching video metadata (not used heavily here)
func CacheSet(ctx context.Context, key string, val string, ttl time.Duration) error {
	return mdb.Redis.Set(ctx, key, val, ttl).Err()
}

func CacheGet(ctx context.Context, key string) (string, error) {
	return mdb.Redis.Get(ctx, key).Result()
}
