package tools

import (
	"log"

	"github.com/jinzhu/gorm"
	"github.com/pangliang/MirServer-Go/orm"
)

var tables = []interface{}{
	&orm.ServerInfo{},
	&orm.User{},
	&orm.Player{},
}

func CreateDatabase(driverName, dataSourceName string, dropTableIfExists bool) {
	db, err := gorm.Open(driverName, dataSourceName)
	if err != nil {
		log.Fatalf("open database error : %s", err)
	}

	if dropTableIfExists {
		for _, table := range tables {
			if err := db.DropTableIfExists(table).Error; err != nil {
				log.Fatal(err)
			}
		}
	}
	for _, table := range tables {
		if err := db.AutoMigrate(table).Error; err != nil {
			log.Fatal(err)
		}
	}
}

func InitDevDB() {
	log.Printf("init dev database ...")

	CreateDatabase("sqlite3", "./mir2.db", true)
	db, err := gorm.Open("sqlite3", "./mir2.db")
	defer db.Close()
	if err != nil {
		log.Fatalf("open database error : %s", err)
	}

	db.Create(&orm.ServerInfo{
		ID:              1,
		GameServerIp:    "192.168.0.166",
		GameServerPort:  7400,
		LoginServerIp:   "192.168.0.166",
		LoginServerPort: 7000,
		Name:            "test1",
	})

	db.Create(&orm.User{
		ID:       1,
		Name:     "11",
		Password: "11",
	})
}

func MigrateDevDB() {
	db, err := gorm.Open("sqlite3", "./mir2.db")
	defer db.Close()
	if err != nil {
		log.Fatalf("open database error : %s", err)
	}

	db.AutoMigrate(tables...)
}


