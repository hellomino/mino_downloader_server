package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"minodl/config"
	"minodl/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	empty         = ""
	keyRds        = "X-RAND-STRING"
	keyTimestamp  = "X-TIMESTAMP"
	KeyClientSign = "X-CLIENT-SIGN"
	keySkipParam  = "X-PARAM-SKIP"
)

// RequestAuthMiddleware 对外服务所有请求合法性验签
func RequestAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// slat为空跳过签名
		salt := config.Get().Slat
		// GET请求不验签
		if salt != "" && c.Request.Method != http.MethodGet && !strings.Contains(c.Request.URL.Path, "/app/callback") {
			requestBody, _ := c.Get("request_body_raw")
			// 请求合法验签
			if !VerifyReqRaw(c.Request, requestBody, salt) {
				log.Printf("ip:%s,req:%s has sign error\n", utils.GetClientIP(c), c.Request.URL.Path)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}
		c.Next()
	}
}

func VerifyReqRaw(r *http.Request, riBody any, salt string) bool {
	// 未配置盐，跳过签名
	if salt == empty {
		return true
	}
	randStr := r.Header.Get(keyRds)
	timestamp := r.Header.Get(keyTimestamp)
	clientSign := r.Header.Get(KeyClientSign)
	skipParam := r.Header.Get(keySkipParam) != empty
	// 必备参数
	if clientSign == empty || timestamp == empty || randStr == empty {
		return false
	}
	paramStr := empty
	bodyBytes, ok := riBody.([]byte)
	if ok && !skipParam {
		paramStr = string(bodyBytes)
		//log.Info("paramStr:", paramStr)
	}
	signString := paramStr + randStr + timestamp + salt
	hash := md5.Sum([]byte(signString))
	serverSign := hex.EncodeToString(hash[:])
	if serverSign == clientSign {
		return true
	}
	log.Printf("client sign: %s, serverSign: %s\n", clientSign, serverSign)
	return false
}
