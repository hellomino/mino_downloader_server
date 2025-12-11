package ws

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"minodl/config"
	"minodl/consts"
	"minodl/log"
	"minodl/utils"
	"minodl/ws/core/connection"
	"minodl/ws/core/message"
	"net/http"
	"strings"
	"time"
)

// Connect 连接websocket
func Connect(c *gin.Context) {
	// 子协议, 至少三个参数
	subs := strings.Split(c.Request.Header.Get("Sec-WebSocket-Protocol"), ",")
	if len(subs) < 3 {
		log.Error("un expected sub protocol %+v", subs)
		return
	}
	// 客户端IP
	clientIP := utils.GetClientIP(c)
	// 防止固定IP恶意连接
	if limited, _ := utils.Limiter.IsLimited(clientIP, time.Second, 5); limited {
		log.Info("limited ip %s create  ws connection, it's too fast.", clientIP)
		return
	}
	// 客户端签名
	clientSign := strings.TrimSpace(subs[2])
	// 每次签名值允许一次连接
	signDone, err := utils.OpLocked(clientSign)
	if err != nil {
		log.Info("opLock err %+v", err)
		return
	}
	defer func() {
		utils.DelOpKey(clientSign)
		signDone()
	}()
	// 验证是否合法
	if salt := config.Get().Slat; salt != consts.EMPTY {
		signString := strings.TrimSpace(subs[0]) + strings.TrimSpace(subs[1]) + salt
		log.Debug("signString: %s, client sign:%s", signString, clientSign)
		hash := md5.Sum([]byte(signString))
		if clientSign != hex.EncodeToString(hash[:]) {
			log.Error("sign not ok sub protocol %+v", subs)
			return
		}
	}
	// 客户端KEY
	clientKey := utils.GetMd5(c.Request.UserAgent() + clientIP)
	// 如果IP和UA一样，未登录情况下KEY是一样的算一个在线
	// clientKey = uuid.New().String() // test code
	// 防止恶意链接
	if limited, _ := utils.Limiter.IsLimited(clientKey, time.Second, 1); limited {
		log.Error("limited ws key %s create  ws connection, it's too fast.", clientIP)
		return
	}
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		Subprotocols:    subs,
	}
	// 升级 HTTP 连接到 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error("WebSocket Upgrader: %v", err)
		return
	}
	defer conn.Close()
	origin := c.Request.Header.Get("Origin")
	log.Debug("client connected,ip:%s, key:%s, origin:%s", clientIP, clientKey, origin)
	h5conn := connection.CreateNewH5Conn(clientKey, origin, conn)
	if h5conn == nil {
		return
	}
	// 从连接中心移除
	defer connection.RemoveConn(h5conn.GetConnectionId())
	// 最多读1kb的消息大小
	conn.SetReadLimit(1024)
	// 处理 WebSocket 消息
	for {
		// 设置读取超时
		_ = conn.SetReadDeadline(time.Now().Add(15 * time.Second))
		// 读取客户端消息
		messageType, rawMessage, err := conn.ReadMessage()
		if err != nil {
			log.Info("client %s connection EOF: %v", clientIP, err)
			break
		}
		_ = conn.SetReadDeadline(time.Time{})
		switch messageType {
		case websocket.TextMessage:
			var msg message.H5Message
			if err = json.Unmarshal(rawMessage, &msg); err != nil {
				log.Error("msg body err: %v, message:%s", err, string(rawMessage))
				break
			}
			execute(h5conn, &msg)
		default:
			log.Error("unknown message type: %s", messageType)
			break
		}
	}
}
