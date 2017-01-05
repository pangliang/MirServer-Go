package loginserver

import (
	"testing"
	"os"
	"github.com/pangliang/MirServer-Go/mockclient"
	"github.com/pangliang/MirServer-Go/protocol"
	"log"
	"path/filepath"
)

const (
	address = "127.0.0.1:7000"
)

func TestMain(m *testing.M) {
	opt := &Option{
		IsTest:true,
		Address:address,
		DbPath:"G:/go_workspace/src/github.com/pangliang/MirServer-Go/mir2.db",
	}
	loginServer := New(opt)
	loginChan := make(chan interface{}, 100)
	loginServer.LoginChan = loginChan
	loginServer.Main()

	retCode := m.Run()
	loginServer.Exit()
	os.Exit(retCode)
}

func TestLogin(t *testing.T) {
	log.Printf("%s\n", filepath.Dir(os.Args[0]))
	client, err := mockclient.New(address)
	defer client.Close()
	if err != nil {
		t.Fatal(err)
	}

	client.Send(&protocol.Packet{protocol.PacketHeader{0, CM_IDPASSWORD, 0, 0, 0}, "11/11"})
	resp, err := client.Read()
	if err != nil {
		t.Fatal(err)
	}
	expect := &protocol.Packet{protocol.PacketHeader{0, SM_PASSOK_SELECTSERVER, 0, 0, 1}, "pangliang-test/1/"}
	if *resp != *expect{
		t.Fatal(resp, expect)
	}
}