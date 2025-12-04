package connection

import (
	"context"
	"github.com/gorilla/websocket"
	"minodl/mdb"
	"sync"
	"sync/atomic"
)

const (
	ConnectedUsers = "connected:users"
	CTypeFree      = "free"
	CTypePaid      = "paid"
)

var (
	NodeOnlineId    int64
	UserConnections = sync.Map{} // 网络连接
)

// CreateNewH5Conn 创建新的H5长连接
func CreateNewH5Conn(userData any, key string, c *websocket.Conn) *H5WsConn {
	redisClient := mdb.Redis
	redisClient.HSet(context.Background(), ConnectedUsers, key, 1)
	hc := &H5WsConn{
		id:       atomic.LoadInt64(&NodeOnlineId),
		key:      key,
		conn:     c,
		UserData: userData,
		Type:     CTypeFree,
	}
	UserConnections.Store(hc.id, hc)
	return hc
}

// RemoveConn 移除连接
func RemoveConn(id int64) {
	UserConnections.Delete(id)
}
