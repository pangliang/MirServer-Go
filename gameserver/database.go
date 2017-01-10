package gameserver

var Tables = []interface{}{
	&Player{},
}

type Player struct {
	Id     uint32
	UserId uint32
	Name   string `gorm:"unique_index"`
	Job    int
	Hair   int
	Level  uint
	Gender int
}
