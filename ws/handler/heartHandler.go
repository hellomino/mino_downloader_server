package handler

import (
	"minodl/log"
	"minodl/ws/core/connection"
	"minodl/ws/core/message"
)

func HeartbeatHandler(conn connection.Conn, msg message.Msg) error {
	if conn.IsLimited(msg) {
		_ = conn.Close()
		return nil
	} else {
		log.Debug("receive heartbeat message:%+v", msg)
		return conn.WriteMessage(msg)
	}
}
