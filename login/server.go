package login

import (
	"github.com/pangliang/MirServer-Go/protocol"
	"github.com/pangliang/go-dao"
	"log"
	_ "github.com/mattn/go-sqlite3"
	"net"
	"bufio"
)

type Session struct {
	Socket net.Conn
}

type LoginServer struct {
	Db *dao.DB
}

func New() *LoginServer {
	db, err := dao.Open("sqlite3", "./mir2.db")
	if err != nil {
		log.Fatalf("open database error : %s", err)
	}

	loginServer := &LoginServer{
		Db:db,
	}

	return loginServer
}

func (l *LoginServer) Handle(socket net.Conn) {
	defer socket.Close()
	session := &Session{
		Socket: socket,
	}
	for {
		reader := bufio.NewReader(socket)
		buf, err := reader.ReadBytes('!')
		if err != nil {
			log.Printf("%v recv err %v", socket.RemoteAddr(), err)
			break
		}
		log.Printf("recv:%s\n", string(buf))

		packet := protocol.Decode(buf)
		log.Printf("packet:%v\n", packet)

		packetHandler, ok := loginHandlers[packet.Header.Protocol]
		if !ok {
			log.Printf("handler not found for protocol : %d \n", packet.Header.Protocol)
			return
		}

		packetHandler(session, packet, l)
	}
}
