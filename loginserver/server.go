package loginserver

import (
	"log"
	"net"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pangliang/MirServer-Go/protocol"
	"github.com/pangliang/MirServer-Go/util"
	"sync/atomic"
	"sync"
)

type env struct {
	sync.RWMutex
	clients         map[int64]*client
	clientIDSequeue int64
}
type Option struct {
	IsTest         bool
	Address        string
	DataSourceName string
	DriverName     string
}

type LoginServer struct {
	env       *env
	opt       *Option
	listener  net.Listener
	waitGroup util.WaitGroupWrapper
}

func New(opt *Option) *LoginServer {
	loginServer := &LoginServer{
		opt: opt,
		env:&env{
			clients : make(map[int64]*client),
		},
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
	packetChan := make(chan *protocol.Packet)

	s.waitGroup.Wrap(func() {
		protocol.PacketPump(socket, packetChan)
	})

	db, err := gorm.Open(s.opt.DriverName, s.opt.DataSourceName)
	if err != nil {
		log.Printf("open database error : %s", err)
		return
	}
	defer db.Close()

	clientId := atomic.AddInt64(&s.env.clientIDSequeue, 1)

	client := &client{
		id: clientId,
		db:     db,
		socket: socket,
		server: s,
		packetChan:packetChan,
	}
	s.env.Lock()
	s.env.clients[clientId] = client
	s.env.Unlock()

	client.Main()
}
