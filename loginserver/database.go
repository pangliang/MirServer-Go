package loginserver

import "github.com/jinzhu/gorm"

type ServerInfo struct {
	Id              uint32
	GameServerIp    string
	GameServerPort  uint32
	LoginServerIp   string
	LoginServerPort uint32
	Name            string
}

type User struct {
	Id     uint32
	Name   string
	Passwd string
	Cert   int32
}

func initDB(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&ServerInfo{},
		&User{},
	).Error; err != nil {
		return err
	}
	return nil
}
