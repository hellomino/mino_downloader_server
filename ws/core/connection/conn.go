package connection

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"minodl/utils"
	"minodl/ws/core/message"
	"minodl/ws/wsmd"
	"strconv"
	"time"
)

type Conn interface {
	GetConnectionId() int64
	WriteMessage(message.Msg) error
	IsLimited(message.Msg) bool
	CheckOrigin(message.Msg) bool
	AddTick()
	SetUser(*wsmd.VPUser)
	GetUser() *wsmd.VPUser
	Close() error
	GetConn() *websocket.Conn
}

type H5WsConn struct {
	id       int64
	key      string
	conn     *websocket.Conn
	UserData *wsmd.VPUser
	Type     string
	Origin   string
	Tick     int
}

func (hw *H5WsConn) GetConn() *websocket.Conn {
	return hw.conn
}

func (hw *H5WsConn) SetUser(u *wsmd.VPUser) {
	hw.UserData = u
}

func (hw *H5WsConn) GetUser() *wsmd.VPUser {
	return hw.UserData
}

func (hw *H5WsConn) AddTick() {
	hw.Tick++
}

func (hw *H5WsConn) CheckOrigin(msg message.Msg) bool {
	return true
}

func (hw *H5WsConn) IsLimited(msg message.Msg) bool {
	// 限制消息发送频率在合理范围
	msgKey := hw.key + strconv.Itoa(msg.GetCode())
	limited, _ := utils.Limiter.IsLimited(msgKey, time.Second, 5)
	if limited {
		_ = hw.WriteMessage(&message.H5Message{
			Code: message.RespError,
			Data: "message limited, please try later",
		})
	}
	return limited
}

func (hw *H5WsConn) Close() error {
	return hw.conn.Close()
}

func (hw *H5WsConn) GetConnectionId() int64 {
	return hw.id
}

func (hw *H5WsConn) WriteMessage(msg message.Msg) error {
	if data, err := json.Marshal(msg); err != nil {
		return err
	} else {
		return hw.conn.WriteMessage(websocket.TextMessage, data)
	}
}
