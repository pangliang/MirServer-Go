package gameserver

import (
	"github.com/pangliang/MirServer-Go/protocol"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"net"
	"bufio"
	"github.com/pangliang/MirServer-Go/util"
	"sync"
	"github.com/pangliang/MirServer-Go/loginserver"
)

type env struct {
	sync.RWMutex
	users map[string]loginserver.User
}

type Option struct {
	IsTest  bool
	Address string
	DbPath  string
}

type Session struct {
	db        *gorm.DB
	attr   map[string]interface{}
	socket net.Conn
}

type GameServer struct {
	opt       *Option
	env       *env
	listener  net.Listener
	waitGroup util.WaitGroupWrapper
	LoginChan <-chan interface{}
	exitChan  chan int
}

func New(opt *Option) *GameServer {
	gameServer := &GameServer{
		opt:opt,
		env:&env{
			users:make(map[string]loginserver.User),
		},
		exitChan:make(chan int),
	}
	return gameServer
}

func (s *GameServer) Main() {
	listener, err := net.Listen("tcp", s.opt.Address)
	if err != nil {
		log.Fatalln("start server error: ", err)
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
			log.Print("exit EventLoop")
			return
		case e := <-s.LoginChan:
			user := e.(loginserver.User)
			s.env.Lock()
			s.env.users[user.Name] = user
			s.env.Unlock()
		}
	}
}

func (s *GameServer) Handle(socket net.Conn) {
	defer socket.Close()
	db, err := gorm.Open("sqlite3", s.opt.DbPath)
	if err != nil {
		log.Fatalf("open database error : %s", err)
	}
	defer db.Close()
	session := &Session{
		db:db,
		socket: socket,
		attr:make(map[string]interface{}),
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

		packetHandler, ok := gameHandlers[packet.Header.Protocol]
		if !ok {
			log.Printf("handler not found for protocol : %d \n", packet.Header.Protocol)
			return
		}

		err = packetHandler(session, packet, s)
		if err != nil {
			log.Printf("handler error: %s\n", err)
		}
	}
}
