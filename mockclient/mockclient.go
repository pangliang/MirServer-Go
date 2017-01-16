package mockclient

import (
	"bufio"
	"log"
	"net"
	"sync/atomic"

	"github.com/pangliang/MirServer-Go/protocol"
	"os"
)

type MockClient struct {
	conn      *net.TCPConn
	reader    *bufio.Reader
	packetSeq uint32
	logger    *log.Logger
}

func New(addr string) (*MockClient, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	client := &MockClient{
		conn:      conn,
		reader:    bufio.NewReader(conn),
		packetSeq: 0,
		logger: log.New(os.Stdout, "", log.Ltime | log.Lshortfile),
	}
	return client, nil
}

func (c *MockClient) Send(p *protocol.Packet) {
	atomic.AddUint32(&c.packetSeq, 1)
	p.SendToServer(c.packetSeq, c.conn)
}

func (c *MockClient) Read() (*protocol.Packet, error) {
	buf, err := c.reader.ReadBytes('!')
	if err != nil {
		return nil, err
	}
	//c.logger.Printf("MockClient recv: %s\n", string(buf))

	packet := protocol.ParseServer(buf)
	c.logger.Printf("packet:%v\n", packet)
	return packet, nil
}

func (c *MockClient) Close() {
	c.conn.Close()
}
