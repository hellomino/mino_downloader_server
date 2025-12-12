package handler

import (
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"minodl/config"
	"minodl/log"
	"minodl/mdb"
	"minodl/utils"
	"minodl/ws/core/connection"
	"minodl/ws/core/message"
	"minodl/ws/wsmd"
	"time"
)

const (
	FREE  = "FREE"
	PRO   = "PRO"
	ULTRA = "ULTRA"
)

func Register(conn connection.Conn, msg message.Msg) error {
	if conn.IsLimited(msg) {
		_ = conn.Close()
	} else {
		var err error
		var plainBytes []byte
		var u wsmd.VPUser
		if plainBytes, err = utils.DecryptString(config.Get().Slat, string(msg.Raw())); err != nil {
			log.Error("decrypt message:%+v, err:%v", msg, err)
		} else {
			var data map[string]string
			if err = json.Unmarshal(plainBytes, &data); err != nil {
				log.Error("Unmarshal message:%+v, err:%v", msg, err)
			} else {
				u = wsmd.VPUser{
					Account:  data["account"],
					Password: data["password"],
					Plan:     PRO,
					Until:    time.Now().AddDate(0, 0, 7).Unix(),
				}
				if err = mdb.Mysql.Create(&u).Error; err != nil {
					log.Error("create user:%+v, err:%v", u, err)
				}
			}
		}
		if err == nil {
			newClue, _ := utils.EncryptString(config.Get().Slat, []byte(u.Account))
			if resp, err := utils.EncryptAny(config.Get().Slat, &map[string]any{
				"id":    u.ID,
				"clue":  newClue,
				"plan":  u.Plan,
				"until": time.Unix(u.Until, 0).Format("2006-01-02 15:04:05"),
			}); err == nil {
				u.Clue = newClue
				conn.SetUser(&u)
				_ = conn.WriteMessage(&message.H5Message{
					Code: message.RespLogin,
					Data: resp,
				})
			} else {
				log.Error("encrypt message:%+v, err:%v", msg, err)
			}
		} else {
			_ = conn.WriteMessage(&message.H5Message{
				Code: message.RespError,
				Data: "server error",
			})
		}
	}
	return nil
}

func Login(conn connection.Conn, msg message.Msg) error {
	if conn.IsLimited(msg) {
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
				if err = mdb.Mysql.Model(&wsmd.VPUser{}).Where("account=?", u.Account).First(&u).Error; err != nil {
					log.Error("find user:%+v, err:%v", u, err)
					msgTip := "server error"
					if errors.Is(err, gorm.ErrRecordNotFound) {
						msgTip = "user not exist, register first"
					}
					_ = conn.WriteMessage(&message.H5Message{
						Code: message.RespError,
						Data: msgTip,
					})
					return err
				} else {
					if u.Password != data["password"] {
						_ = conn.WriteMessage(&message.H5Message{
							Code: message.RespError,
							Data: "password error",
						})
						return err
					}
					newClue, _ := utils.EncryptString(config.Get().Slat, []byte(u.Account))
					if resp, err := utils.EncryptAny(config.Get().Slat, &map[string]any{
						"id":    u.ID,
						"clue":  newClue,
						"plan":  u.Plan,
						"until": time.Unix(u.Until, 0).Format("2006-01-02 15:04:05"),
					}); err == nil {
						u.Clue = newClue
						conn.SetUser(&u)
						_ = conn.WriteMessage(&message.H5Message{
							Code: message.RespLogin,
							Data: resp,
						})
					} else {
						log.Error("encrypt message:%+v, err:%v", msg, err)
					}
				}
			}
		}
	}
	return nil
}
