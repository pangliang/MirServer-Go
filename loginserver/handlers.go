package loginserver

import (
	"github.com/pangliang/MirServer-Go/protocol"
	"log"
	"fmt"
	"math/rand"
	"strings"
	"errors"
)

var loginHandlers = map[uint16]protocol.PacketHandler{
	CM_ADDNEWUSER : func(request *protocol.Packet, args... interface{}) (err error) {
		session := args[0].(*Session)

		params := strings.Split(request.Data, "")
		if len(params) < 4 {
			resp := protocol.NewPacket(SM_NEWID_FAIL)
			resp.Header.Recog = 1
			resp.SendTo(session.socket)
			return
		}
		user := &User{
			Name:strings.Trim(params[1], "\x00"),
			Password:strings.Trim(params[2], "\x00"),
			Cert:0,
		}
		err = session.db.Create(user).Error
		if err != nil {
			resp := protocol.NewPacket(SM_NEWID_FAIL)
			resp.Header.Recog = 2
			resp.SendTo(session.socket)
			return
		}

		resp := protocol.NewPacket(SM_NEWID_SUCCESS)
		resp.SendTo(session.socket)
		return nil
	},
	CM_IDPASSWORD : func(request *protocol.Packet, args... interface{}) (err error) {
		session := args[0].(*Session)
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
		err = session.db.Find(&user, "name=?", params[0]).Error
		if err != nil {
			resp := protocol.NewPacket(SM_PASSWD_FAIL)
			resp.Header.Recog = UserNotFound
			resp.SendTo(session.socket)
			return err
		}

		if user.Password != params[1] {
			resp := protocol.NewPacket(SM_PASSWD_FAIL)
			resp.Header.Recog = WrongPwd
			resp.SendTo(session.socket)
			return errors.New("WrongPwd")
		}

		var serverInfoList []ServerInfo
		err = session.db.Find(&serverInfoList).Error
		if err != nil {
			log.Printf("db list error : %s \n ", err)
			session.socket.Close()
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
		resp.SendTo(session.socket)

		protocol.IOLoop(session.socket, selectServerHandlers, session, &user)

		return nil
	},
}

var selectServerHandlers = map[uint16]protocol.PacketHandler{
	CM_SELECTSERVER : func(request *protocol.Packet, args... interface{}) (err error) {
		session := args[0].(*Session)
		loginUser:= args[1].(*User)

		serverName := request.Data
		var serverInfo ServerInfo
		err = session.db.Find(&serverInfo, "name=?", serverName).Error
		if err != nil {
			resp := &protocol.Packet{}
			resp.Header.Protocol = SM_ID_NOTFOUND
			resp.SendTo(session.socket)
			return
		}

		loginUser.Cert = rand.Int31n(200)
		session.server.LoginChan <- loginUser

		resp := &protocol.Packet{}
		resp.Header.Protocol = SM_SELECTSERVER_OK
		resp.Header.Recog = loginUser.Cert

		resp.Data = fmt.Sprintf("%s/%d/%d",
			serverInfo.GameServerIp,
			serverInfo.GameServerPort,
			loginUser.Cert,
		)

		resp.SendTo(session.socket)

		return nil
	},
}
