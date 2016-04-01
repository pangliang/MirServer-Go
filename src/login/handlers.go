package login

import (
	"core"
	"core/packet"
	"net"
)

var LoginHanders = map[uint16]core.Handler{
	CM_IDPASSWORD : func(request packet.Packet, socket net.Conn) {
		const (
			UserNotFound = 0
			WrongPwd = -1
			WrongPwd3Times = -2
			AlreadyLogin = -3
			NoPay = -4
			BeLock = -5
		)
		resp := new(packet.Packet)
		resp.Header.Protocol = SM_PASSWD_FAIL
		resp.Header.Recog = UserNotFound

		resp.SendTo(socket)
	},
}
