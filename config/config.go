package config

import (
	"encoding/json"
	"flag"
	"os"
)

var cfgFile *string = flag.String("c", "config.json", "config file")

type Config struct {
	MysqlDSN   string `json:"mysqldsn"`
	RedisDSN   string `json:"redisdsn"`
	ListenAddr string `json:"listen_addr"`
	Slat       string `json:"slat"`
	JWTSecret  string `json:"jwt_secret"`
}

var cfg *Config

func Get() Config {
	return *cfg
}

func LoadConfig() (*Config, error) {
	flag.Parse()
	if data, err := os.ReadFile(*cfgFile); err != nil {
		panic(err)
	} else {
		var temp Config
		if err = json.Unmarshal(data, &temp); err != nil {
			panic(err)
		}
		cfg = &temp
	}
	return cfg, nil
}
