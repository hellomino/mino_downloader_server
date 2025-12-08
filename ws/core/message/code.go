package message

const (
	HeartBeat = 10000 + iota // 心跳
	Register
	Login
	LoadPac
	LoadServer
)

const (
	RespRegister = 20000 + iota
	RespLogin
	RespPacScript
	RespServerList
)

const (
	RespError = 30000 + iota
	RespTips
)
