package wsmd

type VPServer struct {
	ID      uint   `gorm:"primaryKey"`
	Host    string `json:"host"`
	Port    string `json:"port"`
	Auth    string `json:"auth"`
	Type    string `json:"type"`
	Icon    string `json:"icon"`
	Name    string `json:"name"`
	Country string `json:"country"`
	Full    bool   `json:"full" gorm:"index:idx_full"`
}
