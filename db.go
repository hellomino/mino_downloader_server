package main

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

func InitGorm(cfg *Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DBDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func InitRedis(cfg *Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})
}

func AutoMigrate(db *gorm.DB) error {
	log.Println("running automigrate...")
	return db.AutoMigrate(&User{}, &Task{})
}
