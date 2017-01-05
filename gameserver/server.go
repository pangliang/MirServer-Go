package gameserver

import (
	"github.com/pangliang/MirServer-Go/protocol"
	"github.com/pangliang/go-dao"
	"log"
	_ "github.com/mattn/go-sqlite3"
	"net"
	"bufio"
	"flag"
	"os"
	"github.com/pangliang/MirServer-Go/util"
	"sync"
)

type user struct {
	username string
	cert int16
}

type env struct {
	sync.RWMutex
	users map[string]*user
}

type Session struct {
	attr   map[string]interface{}
	socket net.Conn
}

type GameServer struct {
	env           *env
	db            *dao.DB
	listener      net.Listener
	waitGroup     util.WaitGroupWrapper
	userLoginChan <-chan map[string]interface{}
	exitChan      chan int
}

func New(userLoginChan <-chan map[string]interface{}) *GameServer {
	db, err := dao.Open("sqlite3", "./mir2.db")
	if err != nil {
		log.Fatalf("open database error : %s", err)
	}

	gameServer := &GameServer{
		db:db,
		userLoginChan:userLoginChan,
		env:&env{
			users:make(map[string]*user),
		},
	}

	return gameServer
}

func (s *GameServer) Main() {
	flagSet := flag.NewFlagSet("gameserver", flag.ExitOnError)
	address := flagSet.String("game-address", "0.0.0.0:7400", "<addr>:<port> to listen on for TCP clients")
	flagSet.Parse(os.Args[1:])

	listener, err := net.Listen("tcp", *address)
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
			break
		case login := <-s.userLoginChan:
			cert := login["cert"].(int16)
			username := login["username"].(string)
			user := &user{username:username, cert:cert}
			s.env.Lock()
			s.env.users[username] = user
			s.env.Unlock()
		}
	}
}

func (l *GameServer) Handle(socket net.Conn) {
	defer socket.Close()
	session := &Session{
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
		log.Printf("recv:%s\n", string(buf))

		packet := protocol.Decode(buf)
		log.Printf("packet:%v\n", packet)

		packetHandler, ok := gameHandlers[packet.Header.Protocol]
		if !ok {
			log.Printf("handler not found for protocol : %d \n", packet.Header.Protocol)
			return
		}

		packetHandler(session, packet, l)
	}
}
