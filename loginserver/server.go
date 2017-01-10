package loginserver

import (
	"log"
	"net"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pangliang/MirServer-Go/protocol"
	"github.com/pangliang/MirServer-Go/util"
)

type Session struct {
	db     *gorm.DB
	socket net.Conn
	server *LoginServer
}

type Option struct {
	IsTest         bool
	Address        string
	DataSourceName string
	DriverName     string
}

type LoginServer struct {
	opt       *Option
	listener  net.Listener
	waitGroup util.WaitGroupWrapper
	LoginChan chan<- interface{}
}

func New(opt *Option) *LoginServer {
	loginServer := &LoginServer{
		opt: opt,
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

func (s *LoginServer) Handle(socket net.Conn) {
	defer socket.Close()

	db, err := gorm.Open(s.opt.DriverName, s.opt.DataSourceName)
	if err != nil {
		log.Printf("open database error : %s", err)
		return
	}
	defer db.Close()

	session := &Session{
		db:     db,
		socket: socket,
		server: s,
	}

	protocol.IOLoop(socket, loginHandlers, session)
}
