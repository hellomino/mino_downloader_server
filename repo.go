package main

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repository struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewRepository(db *gorm.DB, rdb *redis.Client) *Repository {
	return &Repository{db: db, rdb: rdb}
}

func (r *Repository) CreateUser(u *User) error {
	return r.db.Create(u).Error
}

func (r *Repository) GetUserByEmail(email string) (*User, error) {
	var u User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateTask(t *Task) error {
	return r.db.Create(t).Error
}

func (r *Repository) GetTaskByID(id uint) (*Task, error) {
	var t Task
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) ListTasksByUser(userID uint) ([]Task, error) {
	var out []Task
	if err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) UpdateTask(t *Task) error {
	t.UpdatedAt = time.Now()
	return r.db.Save(t).Error
}

// Simple cache helper (JSON could be used). Example usage: caching video metadata (not used heavily here)
func (r *Repository) CacheSet(ctx context.Context, key string, val string, ttl time.Duration) error {
	return r.rdb.Set(ctx, key, val, ttl).Err()
}

func (r *Repository) CacheGet(ctx context.Context, key string) (string, error) {
	return r.rdb.Get(ctx, key).Result()
}
