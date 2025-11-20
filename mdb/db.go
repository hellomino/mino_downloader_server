package mdb

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"minodl/config"
	"minodl/models"
)

var (
	Mysql *gorm.DB
	Redis *redis.Client
)

func InitGorm(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.MysqlDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	Mysql = db
	_ = autoMigrate(db)
	return db, nil
}

func InitRedis(cfg *config.Config) *redis.Client {
	opt, err := redis.ParseURL(cfg.RedisDSN)
	if err != nil {
		log.Fatal(err)
	}
	Redis = redis.NewClient(opt)
	return Redis
}

func autoMigrate(db *gorm.DB) error {
	log.Println("running auto migrate...")
	return db.AutoMigrate(&models.User{}, &models.Task{})
}
