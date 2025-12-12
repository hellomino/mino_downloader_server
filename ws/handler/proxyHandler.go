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

var _ = `
export const MOCK_NODES: ProxyNode[] = [
  { id: '1', name: 'Tokyo Premium 01', flag: '🇯🇵', countryCode: 'JP', latency: 45, host: 'jp1.minoproxy.com', port: 443, type: 'HTTPS' },
  { id: '2', name: 'Singapore Direct', flag: '🇸🇬', countryCode: 'SG', latency: 52, host: 'sg1.minoproxy.com', port: 443, type: 'HTTPS' },
  { id: '3', name: 'US West Fast', flag: '🇺🇸', countryCode: 'US', latency: 120, host: 'us1.minoproxy.com', port: 443, type: 'HTTPS' },
  { id: '4', name: 'Hong Kong CN2', flag: '🇭🇰', countryCode: 'HK', latency: 15, host: 'hk1.minoproxy.com', port: 443, type: 'HTTPS' },
  { id: '5', name: 'London Basic', flag: '🇬🇧', countryCode: 'GB', latency: 180, host: 'uk1.minoproxy.com', port: 443, type: 'HTTPS' },
  { id: '6', name: 'Seoul Gaming', flag: '🇰🇷', countryCode: 'KR', latency: 35, host: 'kr1.minoproxy.com', port: 443, type: 'HTTPS' },
  { id: '7', name: 'Amsterdam P2P', flag: '🇳🇱', countryCode: 'NL', latency: 160, host: 'nl1.minoproxy.com', port: 443, type: 'HTTPS' },
  { id: '8', name: 'Taiwan HiNet', flag: '🇹🇼', countryCode: 'TW', latency: 25, host: 'tw1.minoproxy.com', port: 443, type: 'HTTPS' },
  { id: '9', name: 'Sydney Low Latency', flag: '🇦🇺', countryCode: 'AU', latency: 200, host: 'au1.minoproxy.com', port: 443, type: 'HTTPS' },
  { id: '10', name: 'Toronto Media', flag: '🇨🇦', countryCode: 'CA', latency: 130, host: 'ca1.minoproxy.com', port: 443, type: 'HTTPS' },
  { id: '11', name: 'Paris Secure', flag: '🇫🇷', countryCode: 'FR', latency: 170, host: 'fr1.minoproxy.com', port: 443, type: 'HTTPS' },
  { id: '12', name: 'Mumbai Direct', flag: '🇮🇳', countryCode: 'IN', latency: 250, host: 'in1.minoproxy.com', port: 443, type: 'HTTPS' },
  { id: '13', name: 'Sao Paulo Stream', flag: '🇧🇷', countryCode: 'BR', latency: 280, host: 'br1.minoproxy.com', port: 443, type: 'HTTPS' },
  { id: '14', name: 'Frankfurt Business', flag: '🇩🇪', countryCode: 'DE', latency: 165, host: 'de1.minoproxy.com', port: 443, type: 'HTTPS' },
  { id: '15', name: 'Moscow Private', flag: '🇷🇺', countryCode: 'RU', latency: 190, host: 'ru1.minoproxy.com', port: 443, type: 'HTTPS' },
];
`

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
