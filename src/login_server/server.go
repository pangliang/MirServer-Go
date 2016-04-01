package login_server

import (
	"net"
	"fmt"
	"log"
	"core/packet"
)

const BUFFER_SIZE = 1024;


func handleConnect(socket net.Conn) {
	remain := make([]byte, 0)
	for {
		buf := make([]byte, BUFFER_SIZE)
		size, err := socket.Read(buf)
		if err != nil || size == 0 {
			break
		}
		fmt.Printf("%s\n", string(buf[:size]))

		packets := make([]packet.Packet, 0)
		frames, newRemain := packet.SplitFrame(buf[:size], remain)
		for _, frame := range frames {
			packet := packet.Decode(frame)
			packets = append(packets, packet)
		}
		remain = newRemain

		for _,packet := range packets {
			fmt.Println("packet:%v", packet)
		}
	}
}

func Start(port uint) {
	var listener, err = net.Listen("tcp", fmt.Sprint(":", port))

	if err != nil {
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
			go handleConnect(conn)
		}

	}

}
