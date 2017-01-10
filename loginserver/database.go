package loginserver

var Tables = []interface{}{
	&ServerInfo{},
	&User{},
}

type ServerInfo struct {
	Id              uint32
	Name            string
	LoginServerIp   string
	LoginServerPort uint32
	GameServerIp    string
	GameServerPort  uint32
}

type User struct {
	Id       uint32
	Name     string
	Password string
	Cert     int32
}
