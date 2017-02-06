package orm

type ServerInfo struct {
	ID              uint32
	Name            string
	LoginServerIp   string
	LoginServerPort uint32
	GameServerIp    string
	GameServerPort  uint32
}

type User struct {
	ID              uint32
	Name            string
	Password        string
	Cert            int32
	CurrentServer   ServerInfo `gorm:"ForeignKey:CurrentServerID"`
	CurrentServerID uint32
	Players         []Player
}

type Player struct {
	ID     uint32
	UserId uint32
	Name   string `gorm:"unique_index"`
	Job    int
	Hair   int
	Level  uint
	Gender int
}
