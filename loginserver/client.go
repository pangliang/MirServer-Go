package loginserver

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"net"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/pangliang/MirServer-Go/protocol"
	"github.com/pangliang/MirServer-Go/orm"
)

var ErrNoLogin = errors.New("client no logined")

type client struct {
	id         int64
	db         *gorm.DB
	socket     net.Conn
	server     *LoginServer
	packetChan <-chan *protocol.Packet
	loginUser  *orm.User
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

func (c *client) Exce(packet *protocol.Packet) error {
	switch packet.Header.Protocol {
	case CM_ADDNEWUSER:
		return c.addNewUser(packet)
	case CM_IDPASSWORD:
		return c.idPassword(packet)
	case CM_QUERYCHR:
		return c.queryChr(packet)
	}

	if !c.checkAuth() {
		c.socket.Close()
		return ErrNoLogin
	}

	switch packet.Header.Protocol {
	case CM_SELECTSERVER:
		return c.selectServer(packet)

	case CM_NEWCHR:
		return c.newChr(packet)
	case CM_DELCHR:
		return c.delChr(packet)
	case CM_SELCHR:
		return c.selChr(packet)
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
	user := &orm.User{
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
		UserNotFound   = 0
		WrongPwd       = -1
		WrongPwd3Times = -2
		AlreadyLogin   = -3
		NoPay          = -4
		BeLock         = -5
	)
	params, err := request.Params(2)
	if err != nil {
		return err
	}
	var user orm.User
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

	c.loginUser = &user

	var serverInfoList []orm.ServerInfo
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
		data += fmt.Sprintf("%s/%d/", info.Name, info.ID)
	}
	resp.Data = data
	resp.SendTo(c.socket)

	return nil
}

func (c *client) selectServer(request *protocol.Packet) (err error) {
	if !c.checkAuth() {
		return ErrNoLogin
	}

	serverName := request.Data
	var serverInfo orm.ServerInfo
	err = c.db.Find(&serverInfo, "name=?", serverName).Error
	if err != nil {
		resp := &protocol.Packet{}
		resp.Header.Protocol = SM_ID_NOTFOUND
		resp.SendTo(c.socket)
		return err
	}

	c.loginUser.Cert = rand.Int31n(200)
	c.loginUser.CurrentServerID = serverInfo.ID
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
		serverInfo.LoginServerIp,
		serverInfo.LoginServerPort,
		c.loginUser.Cert,
	)

	resp.SendTo(c.socket)

	return nil
}

func (c *client) queryChr(request *protocol.Packet) (err error) {

	params, err := request.Params(2)
	if err != nil {
		return err
	}

	var user orm.User
	err = c.db.Preload("CurrentServer").Find(&user, "name=?", params[0]).Error
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

	var playerList []orm.Player
	err = c.db.Find(&playerList, &orm.Player{UserId: c.loginUser.ID}).Error
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
		WrongName  = 0
		NameExist  = 2
		MaxPlayers = 3
		SystemErr  = 4
	)

	params, err := request.Params(5)
	if err != nil {
		return err
	}

	player := &orm.Player{
		UserId: c.loginUser.ID,
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

	db := c.db.Delete(orm.Player{}, "user_id=? and name=?", c.loginUser.ID, playerName)
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
	params, err := request.Params(2)
	if err != nil {
		return err
	}

	var player orm.Player
	err = c.db.Find(&player, "user_id=? and name=?", c.loginUser.ID, params[1]).Error
	if err != nil {
		resp := protocol.NewPacket(SM_STARTFAIL)
		resp.Header.Recog = 2
		resp.SendTo(c.socket)
		return err
	}

	log.Printf("selChr:%v", c.loginUser)

	resp := protocol.NewPacket(SM_STARTPLAY)
	resp.Data = fmt.Sprintf("%s/%d",
		c.loginUser.CurrentServer.GameServerIp,
		c.loginUser.CurrentServer.GameServerPort,
	)
	resp.SendTo(c.socket)

	return nil
}

func (c *client) checkAuth() bool {
	return c.loginUser != nil
}
