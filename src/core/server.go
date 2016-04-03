package core

import (
	"net"
	"fmt"
	"log"
	"core/packet"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

const BUFFER_SIZE = 1024;

type Handler func(packet packet.Packet, socket net.Conn, env Env)

type Env struct {
	Db sql.DB
}

type Server struct {
	env      Env
	port     uint32
	handlers map[uint16]Handler
}

func (server *Server) handleConnect(socket net.Conn) {
	defer socket.Close()
	remain := make([]byte, 0)
	for {
		buf := make([]byte, BUFFER_SIZE)
		size, err := socket.Read(buf)
		if err != nil || size == 0 {
			log.Printf("%v recv err %v", socket.RemoteAddr(), err)
			break
		}
		fmt.Printf("recv:%s\n", string(buf[:size]))

		packets := make([]packet.Packet, 0)
		frames, newRemain := packet.SplitFrame(buf[:size], remain)
		for _, frame := range frames {
			packet := packet.Decode(frame)
			packets = append(packets, packet)
		}
		remain = newRemain

		for _, packet := range packets {
			fmt.Printf("packet:%v\n", packet)
			handler, ok := server.handlers[packet.Header.Protocol]
			if ok {
				handler(packet, socket, server.env)
			}else {
				fmt.Printf("handler not found!! \n")
			}
		}
	}
}

func CreateServer(port uint32, handlers map[uint16]Handler) Server{
	db, err := sql.Open("sqlite3", "./mir2.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	env := Env{}
	env.Db = *db

	server := Server{}
	server.env = env
	server.handlers = handlers
	server.port = port

	return server
}

func (server *Server) Start() {

	var listener, err = net.Listen("tcp", fmt.Sprint(":", server.port))

	if err != nil {
		log.Fatalln("start server error: ", err)
	}else {
		defer listener.Close()
		fmt.Println("server start...listening %d", server.port)

		for {
			var conn, err = listener.Accept()
			if err != nil {
				log.Fatal("server accept error:", err)
			}
			fmt.Println("new connect: ", conn.RemoteAddr())
			go server.handleConnect(conn)
		}

	}

}
