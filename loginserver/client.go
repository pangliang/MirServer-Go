package loginserver

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/pangliang/MirServer-Go/protocol"
	"github.com/jinzhu/gorm"
	"net"
)

var ErrNoLogin = errors.New("client no logined")

type client struct {
	id         int64
	db         *gorm.DB
	socket     net.Conn
	server     *LoginServer
	packetChan <-chan *protocol.Packet
	loginUser  *User
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
	case CM_ADDNEWUSER:
		return c.addNewUser(packet)
	case CM_IDPASSWORD:
		return c.idPassword(packet)
	case CM_SELECTSERVER:
		return c.selectServer(packet)
	}
	return errors.New("invalid protocol")
}

func (c *client) addNewUser(request *protocol.Packet) (err error) {
	params := strings.Split(request.Data, "")
	if len(params) < 4 {
		resp := protocol.NewPacket(SM_NEWID_FAIL)
		resp.Header.Recog = 1
		resp.SendTo(c.socket)
		return
	}
	user := &User{
		Name:     strings.Trim(params[1], "\x00"),
		Password: strings.Trim(params[2], "\x00"),
		Cert:     0,
	}
	err = c.db.Create(user).Error
	if err != nil {
		resp := protocol.NewPacket(SM_NEWID_FAIL)
		resp.Header.Recog = 2
		resp.SendTo(c.socket)
		return
	}

	resp := protocol.NewPacket(SM_NEWID_SUCCESS)
	resp.SendTo(c.socket)
	return nil
}

func (c *client) idPassword(request *protocol.Packet) (err error) {
	const (
		UserNotFound = 0
		WrongPwd = -1
		WrongPwd3Times = -2
		AlreadyLogin = -3
		NoPay = -4
		BeLock = -5
	)
	params := request.Params()
	var user User
	err = c.db.Find(&user, "name=?", params[0]).Error
	if err != nil {
		resp := protocol.NewPacket(SM_PASSWD_FAIL)
		resp.Header.Recog = UserNotFound
		resp.SendTo(c.socket)
		return err
	}

	if user.Password != params[1] {
		resp := protocol.NewPacket(SM_PASSWD_FAIL)
		resp.Header.Recog = WrongPwd
		resp.SendTo(c.socket)
		return nil
	}

	c.loginUser = &user;

	var serverInfoList []ServerInfo
	err = c.db.Find(&serverInfoList).Error
	if err != nil {
		c.socket.Close()
		return
	}

	resp := new(protocol.Packet)
	resp.Header.Protocol = SM_PASSOK_SELECTSERVER
	resp.Header.P3 = int16(len(serverInfoList))

	var data string
	for _, info := range serverInfoList {
		data += fmt.Sprintf("%s/%d/", info.Name, info.Id)
	}
	resp.Data = data
	resp.SendTo(c.socket)

	return nil
}

func (c *client) selectServer(request *protocol.Packet) (err error) {
	if (!c.checkAuth()) {
		return ErrNoLogin
	}

	serverName := request.Data
	var serverInfo ServerInfo
	err = c.db.Find(&serverInfo, "name=?", serverName).Error
	if err != nil {
		resp := &protocol.Packet{}
		resp.Header.Protocol = SM_ID_NOTFOUND
		resp.SendTo(c.socket)
		return err
	}

	c.loginUser.Cert = rand.Int31n(200)
	err = c.db.Save(*c.loginUser).Error
	if err != nil {
		resp := &protocol.Packet{}
		resp.Header.Protocol = SM_CERTIFICATION_FAIL
		resp.SendTo(c.socket)
		return err
	}

	resp := &protocol.Packet{}
	resp.Header.Protocol = SM_SELECTSERVER_OK
	resp.Header.Recog = c.loginUser.Cert

	resp.Data = fmt.Sprintf("%s/%d/%d",
		serverInfo.GameServerIp,
		serverInfo.GameServerPort,
		c.loginUser.Cert,
	)

	resp.SendTo(c.socket)

	return nil
}

func (c *client) checkAuth() bool {
	return c.loginUser != nil
}
