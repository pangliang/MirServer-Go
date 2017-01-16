package gameserver

import (
	"log"
	"net"
	"sync"

	"os"
	"sync/atomic"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pangliang/MirServer-Go/protocol"
	"github.com/pangliang/MirServer-Go/util"
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

type GameServer struct {
	logger    *log.Logger
	opt       *Option
	env       *env
	listener  net.Listener
	waitGroup util.WaitGroupWrapper
	exitChan  chan int
}

func New(opt *Option) *GameServer {
	gameServer := &GameServer{
		logger: log.New(os.Stdout, "GameServer:", log.Lshortfile | log.Ltime),
		opt:    opt,
		env: &env{
			clients: make(map[int64]*client),
		},
		exitChan: make(chan int),
	}
	return gameServer
}

func (s *GameServer) Main() {
	listener, err := net.Listen("tcp", s.opt.Address)
	if err != nil {
		s.logger.Fatalln("start server error: ", err)
	}
	s.listener = listener
	s.waitGroup.Wrap(func() {
		protocol.TCPServer(listener, s)
	})

	s.waitGroup.Wrap(func() {
		s.eventLoop()
	})
}

func (s *GameServer) Exit() {
	if s.listener != nil {
		s.listener.Close()
	}
	close(s.exitChan)
	s.waitGroup.Wait()
}

func (s *GameServer) eventLoop() {
	for {
		select {
		case <-s.exitChan:
			s.logger.Print("exit EventLoop")
			return
		}
	}
}

func (s *GameServer) Handle(socket net.Conn) {

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
		id:         clientId,
		db:         db,
		socket:     socket,
		server:     s,
		packetChan: packetChan,
	}

	s.env.Lock()
	s.env.clients[clientId] = client
	s.env.Unlock()

	client.Main()
}
