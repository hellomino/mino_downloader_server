package handler

import (
	"context"
	"encoding/json"
	"minodl/config"
	"minodl/consts"
	"minodl/log"
	"minodl/mdb"
	"minodl/utils"
	"minodl/ws/core/connection"
	"minodl/ws/core/message"
	"minodl/ws/wsmd"
	"strings"
)

func LoadPac(conn connection.Conn, msg message.Msg) error {
	if conn.IsLimited(msg) {
		_ = conn.Close()
	} else {
		if conn.GetUser() == nil {
			_ = conn.WriteMessage(&message.H5Message{
				Code: message.RespTips,
				Data: "please login first",
			})
			return nil
		}
		if resp, err := utils.EncryptString(config.Get().Slat, []byte("pac plain text")); err == nil {
			_ = conn.WriteMessage(&message.H5Message{
				Code: message.RespPacScript,
				Data: resp,
			})
		} else {
			log.Error("encrypt message:%+v, err:%v", msg, err)
		}
	}
	return nil
}

func LoadServer(conn connection.Conn, msg message.Msg) error {
	if conn.IsLimited(msg) {
	} else {
		guest := conn.GetUser() == nil
		clue := config.Get().Slat
		if !guest {
			clue = conn.GetUser().Clue
		}
		var err error
		var plainBytes []byte
		if plainBytes, err = utils.DecryptString(clue, string(msg.Raw())); err != nil {
			log.Error("decrypt message:%+v, err:%v", msg, err)
		} else {
			var data map[string]string
			if err = json.Unmarshal(plainBytes, &data); err != nil {
				log.Error("Unmarshal message:%+v, err:%v", msg, err)
			} else {
				clientIp := data["ip"]
				realIp := strings.Split(conn.GetConn().RemoteAddr().String(), ":")[0]
				if clientIp != realIp {
					log.Error("Ip error, clientIp:%v, realIp:%v", clientIp, realIp)
				} else {
					//TODO pub
				}
				// return server list
				var freeServer, paidServers []wsmd.VPServer
				freeJson := mdb.Redis.Get(context.Background(), consts.FreeServers).Val()
				paidJson := mdb.Redis.Get(context.Background(), consts.PaidServers).Val()
				if freeJson != consts.EMPTY {
					err = json.Unmarshal([]byte(freeJson), &freeServer)
					if err != nil {
						log.Error("Unmarshal message:%+v, err:%v", msg, err)
					}
				}
				if paidJson != consts.EMPTY {
					err = json.Unmarshal([]byte(paidJson), &paidServers)
					if err != nil {
						log.Error("Unmarshal message:%+v, err:%v", msg, err)
					}
				}
				allServers := append(paidServers, freeServer...)
				serversBytes, _ := json.Marshal(&allServers)
				serversBytesEncrypt, _ := utils.EncryptString(clue, serversBytes)
				_ = conn.WriteMessage(&message.H5Message{
					Code: message.RespServerList,
					Data: serversBytesEncrypt,
				})
			}
		}
	}
	return nil
}
