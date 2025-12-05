package handler

import (
	"encoding/json"
	"minodl/config"
	"minodl/log"
	"minodl/mdb"
	"minodl/utils"
	"minodl/ws/core/connection"
	"minodl/ws/core/message"
	"minodl/ws/wsmd"
)

func Register(conn connection.Conn, msg message.Msg) error {
	if conn.IsLimited(msg) {
		_ = conn.Close()
	} else {
		var err error
		var plainBytes []byte
		if plainBytes, err = utils.DecryptString(config.Get().Slat, string(msg.Raw())); err != nil {
			log.Error("decrypt message:%+v, err:%v", msg, err)
		} else {
			var data map[string]string
			if err = json.Unmarshal(plainBytes, &data); err != nil {
				log.Error("Unmarshal message:%+v, err:%v", msg, err)
			} else {
				u := wsmd.VPUser{
					Account:  data["account"],
					Password: data["password"],
				}
				if err = mdb.Mysql.Create(&u).Error; err != nil {
					log.Error("create user:%+v, err:%v", u, err)
				}
			}
		}
		if err == nil {
			if resp, err := utils.EncryptAny(config.Get().Slat, &map[string]string{
				"code": "SUCCESS",
			}); err == nil {
				_ = conn.WriteMessage(&message.H5Message{
					Code: message.RespRegister,
					Data: resp,
				})
			} else {
				log.Error("encrypt message:%+v, err:%v", msg, err)
			}
		}
	}
	return nil
}

func Login(conn connection.Conn, msg message.Msg) error {
	return nil
}

func Logout(conn connection.Conn, msg message.Msg) error {
	return nil
}
