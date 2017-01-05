package loginserver

import (
	"github.com/pangliang/MirServer-Go/protocol"
	"log"
	"fmt"
	"strings"
	"math/rand"
)

type ServerInfo struct {
	Id              uint32
	GameServerIp    string
	GameServerPort  uint32
	LoginServerIp   string
	LoginServerPort uint32
	Name            string
}

var loginHandlers = map[uint16]func(s *Session, request *protocol.Packet, server *LoginServer){
	CM_IDPASSWORD : func(s *Session, request *protocol.Packet, server *LoginServer) {
		const (
			UserNotFound = 0
			WrongPwd = -1
			WrongPwd3Times = -2
			AlreadyLogin = -3
			NoPay = -4
			BeLock = -5
		)
		params := strings.Split(request.Data, "/")
		s.attr["username"] = params[0]
		s.attr["passwd"] = params[1]

		var serverInfoList []ServerInfo
		err := server.db.List(&serverInfoList, "")
		if err != nil {
			log.Printf("db list error : %s \n ", err)
			s.Socket.Close()
			return
		}

		resp := new(protocol.Packet)
		resp.Header.Protocol = SM_PASSOK_SELECTSERVER
		resp.Header.P3 = int16(len(serverInfoList))

		var data string
		for _, info := range serverInfoList {
			data += info.Name + "/" + string(info.Id) + "/"
		}
		resp.Data = data
		resp.SendTo(s.Socket)
	},

	CM_SELECTSERVER : func(s *Session, request *protocol.Packet, server *LoginServer) {

		serverName := request.Data
		var serverInfoList []ServerInfo
		err := server.db.List(&serverInfoList, "where name=?", serverName)
		if err != nil {
			log.Printf("db list error : %s \n ", err)
			s.Socket.Close()
			return
		}

		if len(serverInfoList) == 0 {
			log.Printf("server not found [%s] in {%v}\n", serverName, serverInfoList)
			resp := &protocol.Packet{}
			resp.Header.Protocol = SM_ID_NOTFOUND
			resp.SendTo(s.Socket)
			return
		}

		cert := rand.Int31n(200)
		username := s.attr["username"]
		server.userLoginChan <- map[string]interface{}{
			"username":username,
			"cert":int16(cert),
		}

		resp := &protocol.Packet{}
		resp.Header.Protocol = SM_SELECTSERVER_OK
		resp.Header.Recog = cert

		resp.Data = fmt.Sprintf("%s/%d/%d", serverInfoList[0].GameServerIp, serverInfoList[0].GameServerPort, cert)

		resp.SendTo(s.Socket)
	},
}
