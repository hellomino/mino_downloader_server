# Gin Video Downloader Backend

This is a skeleton Go backend for the Flutter video downloader prototype.

Features:
- Register / Login with JWT
- Create / List / Start / Complete Tasks
- MySQL for persistent storage (GORM)
- Redis for caching
- Placeholder TODOs where actual video parsing/downloading should occur

## Quick start (development)

1. Set environment variables:
```
export MYSQL_DSN='user:pass@tcp(127.0.0.1:3306)/videodb?parseTime=true'
export REDIS_ADDR='127.0.0.1:6379'
export JWT_SECRET='a-very-secret-key'
export PORT=8080
```

2. Run:
```
go mod tidy
go run .
```

## Notes
- The actual per-site video parsing and downloading is intentionally left as TODO. Integrate your parsing/downloader service and call MarkComplete when done.
- For production, secure password storage, HTTPS, rate-limiting, input validation, and more robust background job processing (e.g., worker queue) should be added.
