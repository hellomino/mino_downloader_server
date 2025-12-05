package message

const (
	HeartBeat = 10000 + iota // 心跳
	Register
	Login
	Logout
	LoadServer
)

const (
	RespRegister = 20000 + iota
	RespLogin
	RespLogout
	RespServerList
)
