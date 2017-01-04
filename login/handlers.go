package login

import (
	"github.com/pangliang/MirServer-Go/protocol"
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

		var serverInfoList []ServerInfo
		server.Db.List(&serverInfoList, "")

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
}
