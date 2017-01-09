package loginserver

import (
	"testing"
	"os"
	"github.com/pangliang/MirServer-Go/mockclient"
	"github.com/pangliang/MirServer-Go/protocol"
	"log"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pangliang/MirServer-Go/tools"
)

const (
	SERVER_ADDRESS = "127.0.0.1:7000"
	DB_SOURCE = "g:/go_workspace/src/github.com/pangliang/MirServer-Go/mir2.db"
	DB_DRIVER = "sqlite3"
)

func initTestDB() (err error) {
	tools.CreateDatabase(Tables, DB_DRIVER, DB_SOURCE, true)

	db, err := gorm.Open(DB_DRIVER, DB_SOURCE)
	defer db.Close()
	if err != nil {
		log.Fatalf("open database error : %s", err)
	}

	db.Delete(User{})
	db.Delete(ServerInfo{})

	db.Create(&ServerInfo{
		Id:1,
		GameServerIp:"127.0.0.1",
		GameServerPort:7400,
		LoginServerIp:"127.0.0.1",
		LoginServerPort:7000,
		Name:"test1",
	})

	db.Create(&ServerInfo{
		Id:2,
		GameServerIp:"192.168.0.166",
		GameServerPort:7400,
		LoginServerIp:"192.168.0.166",
		LoginServerPort:7000,
		Name:"test2",
	})

	return nil
}

func TestMain(m *testing.M) {

	err := initTestDB()
	if err != nil {
		log.Fatal(err)
	}

	opt := &Option{
		IsTest:true,
		Address:SERVER_ADDRESS,
		DataSourceName:DB_SOURCE,
		DriverName:DB_DRIVER,
	}
	loginServer := New(opt)
	loginChan := make(chan interface{}, 100)
	loginServer.LoginChan = loginChan
	loginServer.Main()

	retCode := m.Run()
	loginServer.Exit()
	os.Exit(retCode)
}

func sendAndCheck(client *mockclient.MockClient, request *protocol.Packet, expect *protocol.Packet) (err error) {
	client.Send(request)
	resp, err := client.Read()
	if err != nil {
		return
	}
	if *resp != *expect {
		return errors.New(fmt.Sprint(expect, resp))
	}
	return nil
}

func TestCreateUser(t *testing.T) {
	client, err := mockclient.New(SERVER_ADDRESS)
	defer client.Close()
	if err != nil {
		t.Fatal(err)
	}

	if err := sendAndCheck(client,
		&protocol.Packet{protocol.PacketHeader{0, CM_ADDNEWUSER, 0, 0, 0}, "11" + string([]byte{0, 0, 0, 0}) + "11\x00\x00\x00\x00"},
		&protocol.Packet{protocol.PacketHeader{0, SM_NEWID_SUCCESS, 0, 0, 0}, ""},
	); err != nil {
		t.Fatal(err)
	}
}

func TestLogin(t *testing.T) {
	client, err := mockclient.New(SERVER_ADDRESS)
	defer client.Close()
	if err != nil {
		t.Fatal(err)
	}

	if err := sendAndCheck(client,
		&protocol.Packet{protocol.PacketHeader{0, CM_IDPASSWORD, 0, 0, 0}, "11/11"},
		&protocol.Packet{protocol.PacketHeader{0, SM_PASSOK_SELECTSERVER, 0, 0, 2}, "test1/1/test2/2/"},
	); err != nil {
		t.Fatal(err)
	}

	client.Send(&protocol.Packet{protocol.PacketHeader{0, CM_SELECTSERVER, 0, 0, 0}, "test1"})
	resp, err := client.Read()
	if err != nil {
		t.Fatal(fmt.Sprint(err))
	}
	if params := resp.Params(); (len(params) != 3 || params[0] != "127.0.0.1" || params[1] != "7400") ||
		resp.Header.Protocol != SM_SELECTSERVER_OK {
		t.Fatal(fmt.Sprint(resp))
	}

}

func TestLoginFail(t *testing.T) {
	client, err := mockclient.New(SERVER_ADDRESS)
	defer client.Close()
	if err != nil {
		t.Fatal(err)
	}

	if err := sendAndCheck(client,
		&protocol.Packet{protocol.PacketHeader{0, CM_IDPASSWORD, 0, 0, 0}, "111/11"},
		&protocol.Packet{protocol.PacketHeader{0, SM_PASSWD_FAIL, 0, 0, 0}, ""},
	); err != nil {
		t.Fatal(err)
	}

	if err := sendAndCheck(client,
		&protocol.Packet{protocol.PacketHeader{0, CM_IDPASSWORD, 0, 0, 0}, "11/22"},
		&protocol.Packet{protocol.PacketHeader{-1, SM_PASSWD_FAIL, 0, 0, 0}, ""},
	); err != nil {
		t.Fatal(err)
	}
}