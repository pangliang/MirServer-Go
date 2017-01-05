package gameserver

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
	name string
	job Job
	hair byte
	level uint
	gender Gender
}
