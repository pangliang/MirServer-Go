package gameserver

import "github.com/jinzhu/gorm"

type Job int
const(
	/**
	 * 战士
	 */
	Warrior Job = iota
	/**
	 * 法师
	 */
	Wizard
	/**
	 * 道士
	 */
	Taoist
	/**
	 * 英雄
	 */
	Hero
)

type Gender int
const (
	Male Gender = iota
	Female
)

type Player struct {
	Id uint32
	UserId uint32
	Name string `gorm:"unique_index"`
	Job int
	Hair int
	Level uint
	Gender int
}

func initDB(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&Player{},
	).Error; err != nil {
		return err
	}
	return nil
}
