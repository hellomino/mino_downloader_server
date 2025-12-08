package ws

import (
	"minodl/log"
	"minodl/ws/core/connection"
	"minodl/ws/core/message"
	"minodl/ws/handler"
	"runtime/debug"
	"time"
)

var (
	handers = make(map[int]func(conn connection.Conn, msg message.Msg) error)
)

func register(Code int, H func(conn connection.Conn, msg message.Msg) error) {
	handers[Code] = H
}

// 注册需要处理的消息
func init() {
	register(message.HeartBeat, handler.HeartbeatHandler)
	register(message.Register, handler.Register)
	register(message.Login, handler.Login)
	register(message.LoadPac, handler.LoadPac)
	register(message.LoadServer, handler.LoadServer)
}

// 执行处理器，处理消息
func execute(conn connection.Conn, msg message.Msg) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("panic at msg %d handler, stack %s", msg.GetCode(), string(debug.Stack()))
		}
	}()
	begin := time.Now().UnixNano() / int64(time.Millisecond)
	if h, ok := handers[msg.GetCode()]; ok {
		if err := h(conn, msg); err != nil {
			log.Error("handle error at msg %d handler, error %s", msg.GetCode(), err.Error())
		}
	} else {
		log.Error("miss handler error at msg %d", msg.GetCode())
	}
	costs := time.Now().UnixNano()/int64(time.Millisecond) - begin
	log.Debug("===> execute logic %d costs %dms <===", msg.GetCode(), costs)
}
