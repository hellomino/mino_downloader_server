package main

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	DBDSN     string
	RedisAddr string
	RedisPass string
	RedisDB   int
	JWTSecret string
	Port      int
}

func LoadConfigFromEnv() (*Config, error) {
	port := 8080
	if s := os.Getenv("PORT"); s != "" {
		if p, err := strconv.Atoi(s); err == nil {
			port = p
		}
	}
	redisDB := 0
	if s := os.Getenv("REDIS_DB"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			redisDB = v
		}
	}
	jwt := os.Getenv("JWT_SECRET")
	if jwt == "" {
		return nil, errors.New("JWT_SECRET required in env")
	}
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		// example: user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true
		return nil, errors.New("MYSQL_DSN required in env")
	}
	return &Config{
		DBDSN:     dsn,
		RedisAddr: envOr("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPass: envOr("REDIS_PASS", ""),
		RedisDB:   redisDB,
		JWTSecret: jwt,
		Port:      port,
	}, nil
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
