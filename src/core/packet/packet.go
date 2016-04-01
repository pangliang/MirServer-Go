package packet

import (
	"bytes"
	"log"
	"encoding/binary"
	"net"
)

var decode6BitMask = [...]byte{0xfc, 0xf8, 0xf0, 0xe0, 0xc0}

const DEFAULT_PACKET_SIZE = 12;

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

func (packet *Packet) SendTo(socket net.Conn) {
	buf := packet.Encode()
	socket.Write([]byte{'#'})
	socket.Write(buf)
	socket.Write([]byte{'!'})
}

func (packet *Packet) Encode() []byte {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.LittleEndian, packet.Header)
	if err != nil {
		log.Println("Write packet error:", err)
	}
	return append(encoder6BitBuf(buffer.Bytes()), encoder6BitBuf([]byte(packet.Data))...)
}

func Decode(frame []byte) Packet {
	headerFrame := frame[2:DEFAULT_PACKET_SIZE * 4 / 3 + 2]
	packet := Packet{}
	packet.Header.Read(decode6BitBytes(headerFrame))
	packet.Data = string(decode6BitBytes(frame[DEFAULT_PACKET_SIZE * 4 / 3 + 2:]))
	return packet
}

func encoder6BitBuf(src []byte) []byte {
	var size = len(src)
	var destLen = (size / 3) * 4 + 10
	var dest = make([]byte, destLen)
	var destPos = 0
	var resetCount = 0

	var chMade, chRest byte = 0, 0

	for i := 0; i < size; i++ {
		if (destPos >= destLen) {
			break
		}

		chMade = (byte)((chRest | ((src[i] & 0xff) >> uint(2 + resetCount))) & 0x3f)
		chRest = (byte)((((src[i] & 0xff) << uint(8 - (2 + resetCount))) >> uint(2)) & 0x3f)

		resetCount += 2
		if resetCount < 6 {
			dest[destPos] = (byte)(chMade + 0x3c)
			destPos += 1
		}else {
			if (destPos < destLen - 1) {
				dest[destPos] = (byte)(chMade + 0x3c)
				destPos += 1
				dest[destPos] = (byte)(chRest + 0x3c)
				destPos += 1
			}else {
				dest[destPos] = (byte)(chMade + 0x3c)
				destPos += 1
			}

			resetCount = 0;
			chRest = 0;
		}
	}
	if (resetCount > 0 ) {
		dest[destPos] = (byte)(chRest + 0x3c);
		destPos += 1
	}

	dest[destPos] = 0;
	return dest[:destPos];

}

func decode6BitBytes(src []byte) []byte {

	var size = len(src)
	var dest = make([]byte, size * 3 / 4)
	var destPos = 0
	var bitPos uint = 2
	var madeBit uint = 0

	var ch byte = 0
	var chCode byte = 0
	var tmp byte = 0

	for i := 0; i < size; i++ {
		if ((src[i] - 0x3c) >= 0) {
			ch = byte(src[i] - 0x3c)
		} else {
			destPos = 0
			break
		}

		if destPos >= len(dest) {
			break
		}

		if madeBit + 6 >= 8 {
			chCode = byte(tmp | ((ch & 0x3f) >> uint(6 - bitPos)))

			dest[destPos] = chCode
			destPos += 1

			madeBit = 0
			if (bitPos < 6) {
				bitPos += 2
			} else {
				bitPos = 2
				continue
			}
		}

		tmp = (byte)((ch << bitPos) & decode6BitMask[bitPos - 2])

		madeBit += 8 - bitPos
	}

	return dest

}

func SplitFrame(buf []byte, remain []byte) ([][]byte, []byte) {
	frames := make([][]byte, 0)
	offset := 0
	for i, b := range buf {
		if b == '!' {
			frame := buf[offset:i + 1]
			packet := make([]byte, 0, len(remain) + len(frame))
			packet = append(packet, remain...)
			packet = append(packet, frame...)
			remain = remain[len(remain):]
			offset = i + 1
			frames = append(frames, packet)
		}
	}
	remain = append(remain, buf[offset:]...)

	return frames, remain
}