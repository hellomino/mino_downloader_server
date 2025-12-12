package handler

import (
	"minodl/log"
	"minodl/ws/core/connection"
	"minodl/ws/core/message"
)

func HeartbeatHandler(conn connection.Conn, msg message.Msg) error {
	conn.AddTick()
	if conn.IsLimited(msg) {
		return nil
	} else {
		log.Debug("receive heartbeat message:%+v", msg)
		return conn.WriteMessage(msg)
	}
}
