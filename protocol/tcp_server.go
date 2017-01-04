package protocol

import (
	"net"
	"runtime"
	"strings"
	"log"
)


type TCPHandler interface {
	Handle(net.Conn)
}

func TCPServer(listener net.Listener, handler TCPHandler) {
	log.Printf("TCP: listening %s \n", listener.Addr())

	for {
		var conn, err = listener.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				log.Printf("NOTICE: temporary Accept() failure - %s", err)
				runtime.Gosched()
				continue
			}
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Printf("ERROR: listener.Accept() - %s", err)
			}
			break
		}
		log.Println("new connect: ", conn.RemoteAddr())
		go handler.Handle(conn)
	}

	log.Printf("TCP: closing %s \n", listener.Addr())
}


