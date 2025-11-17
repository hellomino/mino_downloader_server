package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func GetClientIP(c *gin.Context) string {
	realIp := c.GetHeader("X-Real-IP")
	if realIp != "" {
		return realIp
	}
	ip := c.GetHeader("X-Forwarded-For")
	if ip != "" {
		return strings.Split(ip, ",")[0]
	}
	// 如果没有 X-Forwarded-For，则尝试 X-Real-IP
	ip = c.GetHeader("X-Real-IP")
	if ip == "" {
		// 如果请求头中都没有，则从 RemoteAddr 获取
		ip = c.ClientIP()
	}
	return ip
}
