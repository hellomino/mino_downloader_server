package wsmd

type VPServer struct {
	ID      uint   `gorm:"primaryKey"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Auth    string `json:"auth"`
	Type    string `json:"type"`
	Icon    string `json:"icon"`
	Name    string `json:"name"`
	Flag    string `json:"flag"`
	Country string `json:"countryCode" gorm:"column:country;index:idx_country"`
	Latency int    `json:"latency"`
	Level   int    `json:"level"`
	Full    bool   `json:"full" gorm:"index:idx_full"`
}
