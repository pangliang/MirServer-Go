package packet

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

func TestSplitPacket(t *testing.T) {
	type Subject struct {
		data    string
		answers []string
	}

	m := []Subject{
		Subject{"1111!2222!33", []string{
			"1111!", "2222!",
		}},
		Subject{"33!4444!", []string{
			"3333!", "4444!",
		}},
		Subject{"55", []string{}},
		Subject{"5", []string{}},
		Subject{"5!6666!77", []string{
			"5555!", "6666!",
		}},
	}
	remain := make([]byte, 0)
	for _, subject := range m {
		packetDatas, newRemain := SplitFrame([]byte(subject.data), remain)
		for i, data := range packetDatas {
			if strings.Compare(string(data), subject.answers[i]) != 0 {
				t.Fatalf("data %s parse wrong : expect:%s[% x], actually:%s[% x]", subject.data, subject.answers[i], subject.answers[i], string(data), string(data))
			}
		}
		remain = newRemain
	}

}
