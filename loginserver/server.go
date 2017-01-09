package loginserver

import (
	"github.com/pangliang/MirServer-Go/protocol"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"net"
	"bufio"
	"github.com/pangliang/MirServer-Go/util"
)

type Session struct {
	db        *gorm.DB
	attr   map[string]interface{}
	Socket net.Conn
}

type Option struct {
	IsTest  bool
	Address string
	DbPath  string
}

type LoginServer struct {
	opt       *Option
	listener  net.Listener
	waitGroup util.WaitGroupWrapper
	LoginChan chan <-interface{}
}

func New(opt *Option) *LoginServer {
	loginServer := &LoginServer{
		opt:opt,
	}
	return loginServer
}

func (s *LoginServer) Main() {
	if s.opt.IsTest {

	}

	listener, err := net.Listen("tcp", s.opt.Address)
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

	db, err := gorm.Open("sqlite3", l.opt.DbPath)
	if err != nil {
		log.Printf("open database error : %s", err)
		return
	}
	defer db.Close()

	session := &Session{
		db:db,
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
		//log.Printf("recv:%s\n", string(buf))

		packet := protocol.ParseClient(buf)
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
