package wsmd

type VPUser struct {
	ID       uint   `gorm:"primaryKey"`
	Account  string `json:"account" gorm:"unique;size:255;not null"`
	Password string `json:"password"`
	Avatar   string `json:"avatar"`
	Until    int64  `json:"until"`
	Paid     bool   `json:"paid"`
}
