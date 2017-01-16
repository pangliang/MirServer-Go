package gameserver

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"log"
	"net"

	"github.com/jinzhu/gorm"
	"github.com/pangliang/MirServer-Go/loginserver"
	"github.com/pangliang/MirServer-Go/protocol"
)

var ErrNoLogin = errors.New("client no logined")

type client struct {
	id         int64
	db         *gorm.DB
	socket     net.Conn
	server     *GameServer
	packetChan <-chan *protocol.Packet
	loginUser  *loginserver.User
}

func (c *client) Main() {
	for {
		select {
		case packet := <-c.packetChan:
			err := c.Exce(packet)
			if err != nil {
				log.Printf("client exec packet %v return err: %v\n", packet, err)
			}
		}
	}
}

func (c *client) Exce(packet *protocol.Packet) (err error) {
	switch packet.Header.Protocol {
	case CM_QUERYCHR:
		return c.queryChr(packet)
	}

	if !c.checkAuth() {
		c.socket.Close()
		return ErrNoLogin
	}

	switch packet.Header.Protocol {
	case CM_NEWCHR:
		return c.newChr(packet)
	case CM_DELCHR:
		return c.delChr(packet)
	case CM_SELCHR:
		return c.selChr(packet)
	}
	return errors.New("invalid protocol")
}

func (c *client) queryChr(request *protocol.Packet) (err error) {

	params := strings.Split(request.Data, "/")
	if len(params) < 2 {
		resp := protocol.NewPacket(SM_QUERYCHR_FAIL)
		resp.Header.Recog = 1
		resp.SendTo(c.socket)
		return
	}

	var user loginserver.User
	err = c.db.Find(&user, "name=?", params[0]).Error
	if err != nil {
		resp := protocol.NewPacket(SM_QUERYCHR_FAIL)
		resp.Header.Recog = 2
		resp.SendTo(c.socket)
		return
	}

	cert, err := strconv.Atoi(params[1])
	if err != nil || int32(cert) != user.Cert {
		resp := protocol.NewPacket(SM_QUERYCHR_FAIL)
		resp.Header.Recog = 3
		resp.SendTo(c.socket)
		return
	}

	c.loginUser = &user

	var playerList []Player
	err = c.db.Find(&playerList, &Player{UserId: c.loginUser.Id}).Error
	if err != nil {
		resp := protocol.NewPacket(SM_QUERYCHR_FAIL)
		resp.Header.Recog = 4
		resp.SendTo(c.socket)
		return
	}

	resp := protocol.NewPacket(SM_QUERYCHR)
	resp.Header.Recog = int32(len(playerList))
	if len(playerList) > 0 {
		for _, player := range playerList {
			resp.Data += fmt.Sprintf("%s/%d/%d/%d/%d/",
				player.Name,
				player.Job,
				player.Hair,
				player.Level,
				player.Gender,
			)
		}
	}
	resp.SendTo(c.socket)

	return nil
}

func (c *client) newChr(request *protocol.Packet) (err error) {

	const (
		WrongName = 0
		NameExist = 2
		MaxPlayers = 3
		SystemErr = 4
	)

	params := strings.Split(request.Data, "/")
	if len(params) < 5 {
		resp := protocol.NewPacket(SM_NEWCHR_FAIL)
		resp.Header.Recog = 1
		resp.SendTo(c.socket)
		return nil
	}

	player := &Player{
		UserId: c.loginUser.Id,
		Level:  1,
	}
	player.Name = params[1]
	player.Hair, _ = strconv.Atoi(params[2])
	player.Job, _ = strconv.Atoi(params[3])
	player.Gender, _ = strconv.Atoi(params[4])

	err = c.db.Create(player).Error
	if err != nil {
		resp := protocol.NewPacket(SM_NEWCHR_FAIL)
		resp.Header.Recog = NameExist
		resp.SendTo(c.socket)
		return err
	}

	protocol.NewPacket(SM_NEWCHR_SUCCESS).SendTo(c.socket)
	return nil
}

func (c *client) delChr(request *protocol.Packet) (err error) {

	playerName := request.Data

	db := c.db.Delete(Player{}, "user_id=? and name=?", c.loginUser.Id, playerName)
	if db.Error != nil || db.RowsAffected != 1 {
		resp := protocol.NewPacket(SM_DELCHR_FAIL)
		resp.Header.Recog = 2
		resp.SendTo(c.socket)
		return db.Error
	}
	resp := protocol.NewPacket(SM_DELCHR_SUCCESS)
	resp.SendTo(c.socket)

	return nil
}

func (c *client) selChr(request *protocol.Packet) (err error) {
	params := request.Params()
	if len(params) != 2 {
		resp := protocol.NewPacket(SM_STARTFAIL)
		resp.Header.Recog = 1
		resp.SendTo(c.socket)
		return errors.New("params wrong")
	}

	var player Player
	err = c.db.Find(&player, "user_id=? and name=?", c.loginUser.Id, params[1]).Error
	if err != nil {
		resp := protocol.NewPacket(SM_STARTFAIL)
		resp.Header.Recog = 2
		resp.SendTo(c.socket)
		return err
	}

	resp := protocol.NewPacket(SM_STARTPLAY)
	resp.SendTo(c.socket)

	return nil
}

func (c *client) checkAuth() bool {
	return (c.loginUser != nil)
}
