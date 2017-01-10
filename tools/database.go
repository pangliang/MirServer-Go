package tools

import (
	"log"

	"github.com/jinzhu/gorm"
)

func CreateDatabase(tables []interface{}, driverName, dataSourceName string, dropTableIfExists bool) {
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
