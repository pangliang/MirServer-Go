package gameserver

import (
	"github.com/pangliang/MirServer-Go/protocol"
	"strings"
	"strconv"
	"fmt"
)

var gameHandlers = map[uint16]func(session *Session, request *protocol.Packet, server *GameServer){
	CM_QUERYCHR : func(session *Session, request *protocol.Packet, server *GameServer) {
		params := strings.Split(request.Data, "/")
		if len(params) < 1 {
			protocol.NewPacket(SM_CERTIFICATION_FAIL).SendTo(session.socket)
			return
		}
		username := params[0]
		server.env.RLock()
		loginUser, ok := server.env.users[username]
		server.env.RUnlock()

		if !ok {
			protocol.NewPacket(SM_CERTIFICATION_FAIL).SendTo(session.socket)
			return
		}

		cert, err := strconv.Atoi(params[1])
		if err != nil || int16(cert) != loginUser.cert {
			protocol.NewPacket(SM_CERTIFICATION_FAIL).SendTo(session.socket)
			return
		}
		player := &Player{
			name:"pangliang",
			job:Warrior,
			hair:0,
			level:1,
			gender:Female,
		}

		session.attr["user"] = loginUser
		resp := protocol.NewPacket(SM_QUERYCHR)
		resp.Header.Recog = 1
		resp.Data = fmt.Sprintf("%s/%d/%d/%d/%d/",
			player.name,
			player.job,
			player.hair,
			player.level,
			player.gender,
		)
		resp.SendTo(session.socket)

		return
	},
}
