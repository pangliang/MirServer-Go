package protocol

import (
	"fmt"
	"strings"
	"testing"
)

func packetEquel(p1 Packet, p2 Packet) bool {
	return strings.Compare(fmt.Sprintf("%v", p1), fmt.Sprintf("%v", p2)) == 0
}

func TestParsePackets(t *testing.T) {
	type Subject struct {
		data    string
		answers Packet
	}
	m := []Subject{
		Subject{"#1<<<<<I@C<<<<<<<<TNy]!", Packet{PacketHeader{0, 2001, 0, 0, 0}, "a/a"}},
	}
	for _, subject := range m {
		data := []byte(subject.data)
		packet := Decode(data)
		fmt.Printf("%v\n", packet)
		if ! packetEquel(packet, subject.answers) {
			t.Fatalf("Decode fata !! expect: %v actually:%v", subject.answers, packet)
		}

		encoded := string(subject.answers.Encode())

		if subject.data[2:len(subject.data) -1 ] != encoded {
			t.Fatalf("Encode fata !! expect: %v actually:%v", subject.data, encoded)
		}
	}
}
