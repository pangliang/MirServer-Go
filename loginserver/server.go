package loginserver

import (
	"github.com/pangliang/MirServer-Go/protocol"
	"github.com/pangliang/go-dao"
	"log"
	_ "github.com/mattn/go-sqlite3"
	"net"
	"bufio"
	"github.com/pangliang/MirServer-Go/util"
	"flag"
	"os"
)

type Session struct {
	attr   map[string]interface{}
	Socket net.Conn
}

type LoginServer struct {
	db         *dao.DB
	listener   net.Listener
	waitGroup  util.WaitGroupWrapper
	loginChan  chan <-  interface{}
	packetChan chan *protocol.Packet
}

func New(loginChan chan <- interface{}) *LoginServer {
	db, err := dao.Open("sqlite3", "./mir2.db")
	if err != nil {
		log.Fatalf("open database error : %s", err)
	}

	loginServer := &LoginServer{
		db:db,
		loginChan: loginChan,
		packetChan:make(chan *protocol.Packet, 1),
	}

	return loginServer
}

func (s *LoginServer) Main() {
	flagSet := flag.NewFlagSet("loginserver", flag.ExitOnError)
	address := flagSet.String("login-address", "0.0.0.0:7000", "<addr>:<port> to listen on for TCP clients")
	flagSet.Parse(os.Args[1:])

	listener, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatalln("start server error: ", err)
	}
	s.listener = listener
	s.waitGroup.Wrap(func() {
		protocol.TCPServer(listener, s)
	})
}

func (s *LoginServer) Exit() {
	if s.listener != nil {
		s.listener.Close()
	}
	s.waitGroup.Wait()
}

func (l *LoginServer) Handle(socket net.Conn) {
	defer socket.Close()
	session := &Session{
		Socket: socket,
		attr:map[string]interface{}{},
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

		err = packetHandler(session, packet, l)
		if err != nil {
			log.Printf("handler error: %s\n", err)
		}
	}
}
