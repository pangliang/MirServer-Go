package loginserver

import (
	"github.com/pangliang/MirServer-Go/protocol"
	"log"
	"fmt"
	"math/rand"
	"strings"
)

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

var loginHandlers = map[uint16]func(s *Session, request *protocol.Packet, server *LoginServer) (err error){
	CM_ADDNEWUSER : func(session *Session, request *protocol.Packet, server *LoginServer) (err error) {
		params := strings.Split(request.Data, "")
		if len(params) < 4 {
			resp := protocol.NewPacket(SM_NEWID_FAIL)
			resp.Header.Recog = 1
			resp.SendTo(session.Socket)
			return
		}
		user := &User{
			Name:strings.Trim(params[1], "\x00"),
			Passwd:strings.Trim(params[2], "\x00"),
			Cert:0,
		}
		_, err = server.db.Save(*user)
		if err != nil {
			resp := protocol.NewPacket(SM_NEWID_FAIL)
			resp.Header.Recog = 2
			resp.SendTo(session.Socket)
			return
		}

		resp := protocol.NewPacket(SM_NEWID_SUCCESS)
		resp.SendTo(session.Socket)
		return nil
	},
	CM_IDPASSWORD : func(session *Session, request *protocol.Packet, server *LoginServer) (err error) {
		const (
			UserNotFound = 0
			WrongPwd = -1
			WrongPwd3Times = -2
			AlreadyLogin = -3
			NoPay = -4
			BeLock = -5
		)
		params := request.Params()
		var userList []User
		err = server.db.List(&userList, "where name=?", params[0])
		if err != nil || len(userList) == 0 {
			resp := protocol.NewPacket(SM_PASSWD_FAIL)
			resp.Header.Recog = UserNotFound
			resp.SendTo(session.Socket)
			return
		}

		if userList[0].Passwd != params[1] {
			resp := protocol.NewPacket(SM_PASSWD_FAIL)
			resp.Header.Recog = WrongPwd
			resp.SendTo(session.Socket)
			return
		}

		session.attr["user"] = userList[0]

		var serverInfoList []ServerInfo
		err = server.db.List(&serverInfoList, "")
		if err != nil {
			log.Printf("db list error : %s \n ", err)
			session.Socket.Close()
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
		resp.SendTo(session.Socket)

		return nil
	},

	CM_SELECTSERVER : func(s *Session, request *protocol.Packet, server *LoginServer) (err error) {

		serverName := request.Data
		var serverInfoList []ServerInfo
		err = server.db.List(&serverInfoList, "where name=?", serverName)
		if err != nil || len(serverInfoList) == 0 {
			resp := &protocol.Packet{}
			resp.Header.Protocol = SM_ID_NOTFOUND
			resp.SendTo(s.Socket)
			return
		}

		user := s.attr["user"].(User)
		user.Cert = rand.Int31n(200)
		server.LoginChan <- user

		resp := &protocol.Packet{}
		resp.Header.Protocol = SM_SELECTSERVER_OK
		resp.Header.Recog = user.Cert

		resp.Data = fmt.Sprintf("%s/%d/%d",
			serverInfoList[0].GameServerIp,
			serverInfoList[0].GameServerPort,
			user.Cert,
		)

		resp.SendTo(s.Socket)

		return nil
	},
}
