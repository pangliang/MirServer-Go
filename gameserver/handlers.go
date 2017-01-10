package gameserver

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/pangliang/MirServer-Go/loginserver"
	"github.com/pangliang/MirServer-Go/protocol"
)

var gameLoginHandler = map[uint16]protocol.PacketHandler{
	CM_QUERYCHR: func(request *protocol.Packet, args ...interface{}) (err error) {
		session := args[0].(*Session)

		params := strings.Split(request.Data, "/")
		if len(params) < 2 {
			resp := protocol.NewPacket(SM_QUERYCHR_FAIL)
			resp.Header.Recog = 1
			resp.SendTo(session.socket)
			return
		}
		username := params[0]
		session.server.env.RLock()
		loginUser, ok := session.server.env.users[username]
		session.server.env.RUnlock()

		if !ok {
			resp := protocol.NewPacket(SM_QUERYCHR_FAIL)
			resp.Header.Recog = 2
			resp.SendTo(session.socket)
			return
		}

		cert, err := strconv.Atoi(params[1])
		if err != nil || int32(cert) != loginUser.Cert {
			resp := protocol.NewPacket(SM_QUERYCHR_FAIL)
			resp.Header.Recog = 3
			resp.SendTo(session.socket)
			return
		}

		var playerList []Player
		err = session.db.Find(&playerList, &Player{UserId: loginUser.Id}).Error
		if err != nil {
			resp := protocol.NewPacket(SM_QUERYCHR_FAIL)
			resp.Header.Recog = 4
			resp.SendTo(session.socket)
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
		resp.SendTo(session.socket)

		protocol.IOLoop(session.socket, playerHandlers, session, loginUser)
		return nil
	},
}

var playerHandlers = map[uint16]protocol.PacketHandler{
	CM_NEWCHR: func(request *protocol.Packet, args ...interface{}) (err error) {
		session := args[0].(*Session)
		loginUser := args[1].(*loginserver.User)
		const (
			WrongName  = 0
			NameExist  = 2
			MaxPlayers = 3
			SystemErr  = 4
		)
		params := strings.Split(request.Data, "/")
		if len(params) < 5 {
			resp := protocol.NewPacket(SM_NEWCHR_FAIL)
			resp.Header.Recog = 1
			resp.SendTo(session.socket)
			return nil
		}

		player := &Player{
			UserId: loginUser.Id,
			Level:  1,
		}
		player.Name = params[1]
		player.Hair, _ = strconv.Atoi(params[2])
		player.Job, _ = strconv.Atoi(params[3])
		player.Gender, _ = strconv.Atoi(params[4])

		err = session.db.Create(player).Error
		if err != nil {
			resp := protocol.NewPacket(SM_NEWCHR_FAIL)
			resp.Header.Recog = NameExist
			resp.SendTo(session.socket)
			return err
		}

		protocol.NewPacket(SM_NEWCHR_SUCCESS).SendTo(session.socket)
		return nil
	},
	CM_QUERYCHR: func(request *protocol.Packet, args ...interface{}) (err error) {
		session := args[0].(*Session)
		loginUser := args[1].(*loginserver.User)

		var playerList []Player
		err = session.db.Find(&playerList, &Player{UserId: loginUser.Id}).Error
		if err != nil {
			resp := protocol.NewPacket(SM_QUERYCHR_FAIL)
			resp.Header.Recog = 4
			resp.SendTo(session.socket)
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
		resp.SendTo(session.socket)
		return nil
	},
	CM_DELCHR: func(request *protocol.Packet, args ...interface{}) (err error) {
		session := args[0].(*Session)
		loginUser := args[1].(*loginserver.User)
		playerName := request.Data

		db := session.db.Delete(Player{}, "user_id=? and name=?", loginUser.Id, playerName)
		if db.Error != nil || db.RowsAffected != 1 {
			resp := protocol.NewPacket(SM_DELCHR_FAIL)
			resp.Header.Recog = 2
			resp.SendTo(session.socket)
			return db.Error
		}
		resp := protocol.NewPacket(SM_DELCHR_SUCCESS)
		resp.SendTo(session.socket)

		return nil
	},
	CM_SELCHR: func(request *protocol.Packet, args ...interface{}) (err error) {
		session := args[0].(*Session)
		loginUser := args[1].(*loginserver.User)

		params := request.Params()
		if len(params) != 2 {
			resp := protocol.NewPacket(SM_STARTFAIL)
			resp.Header.Recog = 1
			resp.SendTo(session.socket)
			return errors.New("params wrong")
		}

		var player Player
		err = session.db.Find(&player, "user_id=? and name=?", loginUser.Id, params[1]).Error
		if err != nil {
			resp := protocol.NewPacket(SM_STARTFAIL)
			resp.Header.Recog = 2
			resp.SendTo(session.socket)
			return err
		}

		resp := protocol.NewPacket(SM_STARTPLAY)
		resp.SendTo(session.socket)

		return nil
	},
}
