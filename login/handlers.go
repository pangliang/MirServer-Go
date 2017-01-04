package login

import (
	"net"
	"github.com/pangliang/MirServer-Go/core"
	"github.com/pangliang/MirServer-Go/core/packet"
)

type ServerInfo struct {
	Id              uint32
	GameServerIp    string
	GameServerPort  uint32
	LoginServerIp   string
	LoginServerPort uint32
	Name            string
}

var LoginHanders = map[uint16]core.Handler{
	CM_IDPASSWORD : func(request packet.Packet, socket net.Conn, env core.Env) {
		const (
			UserNotFound = 0
			WrongPwd = -1
			WrongPwd3Times = -2
			AlreadyLogin = -3
			NoPay = -4
			BeLock = -5
		)

		var serverInfoList []ServerInfo
		env.Db.List(&serverInfoList, "")

		resp := new(packet.Packet)
		resp.Header.Protocol = SM_PASSOK_SELECTSERVER
		resp.Header.P3 = int16(len(serverInfoList))

		var data string
		for _, info := range serverInfoList {
			data += info.Name + "/" + string(info.Id) + "/"
		}
		resp.Data = data
		resp.SendTo(socket)
	},
}
