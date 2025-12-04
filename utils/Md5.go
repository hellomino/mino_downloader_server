package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
)

func GetMd5(code string) string {
	// 获取md5
	hash := md5.New()
	_, _ = io.WriteString(hash, code)
	return hex.EncodeToString(hash.Sum(nil))
}
