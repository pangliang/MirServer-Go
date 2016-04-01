package core

import (
	"net"
	"fmt"
	"log"
	"core/packet"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"path/filepath"
)

const BUFFER_SIZE = 1024;

type Handler func(packet packet.Packet, socket net.Conn)

type Server struct {
	Port     uint32
	Handlers map[uint16] Handler
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

		for _,packet := range packets {
			fmt.Printf("packet:%v\n", packet)
			handler, ok := server.Handlers[packet.Header.Protocol]
			if ok {
				handler(packet, socket)
			}else{
				fmt.Printf("handler not found!! \n")
			}
		}
	}
}

func (server *Server) Start() {

	db, err := sql.Open("sqlite3", "./mir2.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var listener, err2 = net.Listen("tcp", fmt.Sprint(":", server.Port))

	if err2 != nil {
		log.Fatalln("start server error: ", err)
	}else {
		defer listener.Close()
		fmt.Println("server start...")

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
