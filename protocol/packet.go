package protocol

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strings"
	"errors"
)

var decode6BitMask = [...]byte{0xfc, 0xf8, 0xf0, 0xe0, 0xc0}

const (
	DEFAULT_PACKET_SIZE = 12
	CONTENT_SEPARATOR   = "/"
)

var ERR_FORMAT_WRONG = errors.New("Packet format wrong")
type PacketHeader struct {
	Recog    int32
	Protocol uint16
	P1       int16
	P2       int16
	P3       int16
}

func (header *PacketHeader) Read(buf []byte) {
	buffer := bytes.NewBuffer(buf)
	err := binary.Read(buffer, binary.LittleEndian, header)
	if err != nil {
		log.Println("decode packet error:", err)
	}
}

type Packet struct {
	Header PacketHeader
	Data   string
}

type PacketHandler func(packet *Packet, args ...interface{}) error

func PacketPump(socket net.Conn, packetChan chan<- *Packet) {
	reader := bufio.NewReader(socket)
	for {
		buf, err := reader.ReadBytes('!')
		if err != nil {
			log.Printf("%v recv err %v", socket.RemoteAddr(), err)
			return
		}
		//log.Printf("recv:%s\n", string(buf))

		packet := ParseClient(buf)
		log.Printf("recv:%v\n", packet)

		packetChan <- packet
	}
}

func NewPacket(protocolId uint16) *Packet {
	p := &Packet{}
	p.Header.Protocol = protocolId
	return p
}

func (packet *Packet) Params(least int) ([]string, error) {
	params := strings.Split(packet.Data, CONTENT_SEPARATOR)
	if len(params) < least {
		return nil, ERR_FORMAT_WRONG
	}
	return params, nil
}

func (packet *Packet) SendTo(socket net.Conn) {
	log.Printf("send:%v\n", packet)
	data := packet.encode()
	socket.Write([]byte{'#'})
	socket.Write([]byte(data))
	socket.Write([]byte{'!'})
}

func (packet *Packet) SendToServer(seq uint32, socket net.Conn) {
	log.Printf("sendToServ:%v\n", packet)
	buf := packet.encode()
	data := fmt.Sprintf("#%d%s!", seq, buf)
	socket.Write([]byte(data))
}

func (packet *Packet) encode() []byte {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.LittleEndian, packet.Header)
	if err != nil {
		log.Println("Write packet error:", err)
	}
	return append(encoder6BitBuf(buffer.Bytes()), encoder6BitBuf([]byte(packet.Data))...)
}

func decode(frame []byte) *Packet {
	decodeFrame := decode6BitBytes(frame)
	//log.Printf("decoce:%s", string(decodeFrame))
	packet := &Packet{}
	if decodeFrame[0] == byte('*') && decodeFrame[1] == byte('*') {
		// [**11/2222/81/20020522/9]  game login packet
		packet.Header.Protocol = 65001
		packet.Data = string(decodeFrame[2:])
	} else {
		packet.Header.Read(decodeFrame[:DEFAULT_PACKET_SIZE])
		packet.Data = string(decodeFrame[DEFAULT_PACKET_SIZE:])
	}

	return packet
}

func ParseClient(frame []byte) *Packet {
	return decode(frame[2 : len(frame)-1])
}

func ParseServer(frame []byte) *Packet {
	return decode(frame[1 : len(frame)-1])
}

func encoder6BitBuf(src []byte) []byte {
	var size = len(src)
	var destLen = (size/3)*4 + 10
	var dest = make([]byte, destLen)
	var destPos = 0
	var resetCount = 0

	var chMade, chRest byte = 0, 0

	for i := 0; i < size; i++ {
		if destPos >= destLen {
			break
		}

		chMade = (byte)((chRest | ((src[i] & 0xff) >> uint(2+resetCount))) & 0x3f)
		chRest = (byte)((((src[i] & 0xff) << uint(8-(2+resetCount))) >> uint(2)) & 0x3f)

		resetCount += 2
		if resetCount < 6 {
			dest[destPos] = (byte)(chMade + 0x3c)
			destPos += 1
		} else {
			if destPos < destLen-1 {
				dest[destPos] = (byte)(chMade + 0x3c)
				destPos += 1
				dest[destPos] = (byte)(chRest + 0x3c)
				destPos += 1
			} else {
				dest[destPos] = (byte)(chMade + 0x3c)
				destPos += 1
			}

			resetCount = 0
			chRest = 0
		}
	}
	if resetCount > 0 {
		dest[destPos] = (byte)(chRest + 0x3c)
		destPos += 1
	}

	dest[destPos] = 0
	return dest[:destPos]

}

func decode6BitBytes(src []byte) []byte {

	var size = len(src)
	var dest = make([]byte, size*3/4)
	var destPos = 0
	var bitPos uint = 2
	var madeBit uint = 0

	var ch byte = 0
	var chCode byte = 0
	var tmp byte = 0

	for i := 0; i < size; i++ {
		if (src[i] - 0x3c) >= 0 {
			ch = byte(src[i] - 0x3c)
		} else {
			destPos = 0
			break
		}

		if destPos >= len(dest) {
			break
		}

		if madeBit+6 >= 8 {
			chCode = byte(tmp | ((ch & 0x3f) >> uint(6-bitPos)))

			dest[destPos] = chCode
			destPos += 1

			madeBit = 0
			if bitPos < 6 {
				bitPos += 2
			} else {
				bitPos = 2
				continue
			}
		}

		tmp = (byte)((ch << bitPos) & decode6BitMask[bitPos-2])

		madeBit += 8 - bitPos
	}

	return dest
}
