package gameserver

import (
	"errors"

	"log"
	"net"

	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/pangliang/MirServer-Go/orm"
	"github.com/pangliang/MirServer-Go/protocol"
)

var ErrNoLogin = errors.New("client no logined")

type client struct {
	id          int64
	db          *gorm.DB
	socket      net.Conn
	server      *GameServer
	requestChan <-chan *protocol.Packet
	player      *orm.Player
}

func (c *client) Main() {
	for {
		select {
		case request := <-c.requestChan:
			err := c.Exce(request)
			if err != nil {
				log.Printf("client exec packet %v return err: %v\n", request, err)
			}
		}
	}
}

func (c *client) Exce(request *protocol.Packet) (err error) {
	switch request.Header.Protocol {
	case CM_GAMELOGIN:
		return c.gameLogin(request)
	}

	if !c.checkAuth() {
		c.socket.Close()
		return ErrNoLogin
	}

	switch request.Header.Protocol {
	case CM_LOGINNOTICEOK:
		return c.loginNoticeOK(request)
	}
	return errors.New("invalid protocol")
}

func (c *client) gameLogin(request *protocol.Packet) (err error) {
	params, err := request.Params(5)
	if err != nil {
		return err
	}

	var user orm.User
	err = c.db.Preload("Players").Find(&user, "name=?", params[0]).Error
	if err != nil {
		resp := protocol.NewPacket(SM_CERTIFICATION_FAIL)
		resp.Header.Recog = 2
		resp.SendTo(c.socket)
		return
	}

	cert, err := strconv.Atoi(params[2])
	if err != nil || int32(cert) != user.Cert {
		resp := protocol.NewPacket(SM_CERTIFICATION_FAIL)
		resp.SendTo(c.socket)
		return
	}

	for _, player := range user.Players {
		if player.Name == params[1] {
			c.player = &player
			resp := protocol.NewPacket(SM_SENDNOTICE)
			notice := "======\\\\111111\\\\2222"
			resp.Header.Recog = int32(len(notice))
			resp.Data = notice
			resp.SendTo(c.socket)
			return nil
		}
	}
	return nil
}

func (c *client) loginNoticeOK(request *protocol.Packet) (err error) {
	logon := protocol.NewPacket(SM_LOGON)
	logon.Header.Recog = 1000
	logon.Header.P1 = 273
	logon.Header.P2 = 590
	logon.Header.P3 = ((1 << 8) | 1)
	logon.Data = string([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	logon.SendTo(c.socket)
	return nil
}

func (c *client) checkAuth() bool {
	return (c.player != nil)
}
