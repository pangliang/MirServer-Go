package gameserver

import (
	"testing"
	"os"
	"github.com/pangliang/MirServer-Go/mockclient"
	"github.com/pangliang/MirServer-Go/protocol"
	"github.com/jinzhu/gorm"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"errors"
	"fmt"
	"github.com/pangliang/MirServer-Go/loginserver"
	"strconv"
)

const (
	LOGIN_SERVER_ADDRESS = "127.0.0.1:7000"
	GAME_SERVER_ADDRESS = "127.0.0.1:7400"
	DB_PATH = "g:/go_workspace/src/github.com/pangliang/MirServer-Go/mir2.db"
)

func initTestDB() (err error) {
	db, err := gorm.Open("sqlite3", DB_PATH)
	if err != nil {
		log.Fatalf("open database error : %s", err)
	}
	err = initDB(db)
	if err != nil {
		log.Fatalln("init database error: ", err)
	}

	db.Delete(loginserver.User{})
	db.Create(&loginserver.User{
		Id:1,
		Name:"pangliang",
		Passwd:"pwd",
	})

	db.Delete(loginserver.ServerInfo{})
	db.Create(&loginserver.ServerInfo{
		Id:1,
		GameServerIp:"127.0.0.1",
		GameServerPort:7400,
		LoginServerIp:"127.0.0.1",
		LoginServerPort:7000,
		Name:"test1",
	})

	db.Create(&loginserver.ServerInfo{
		Id:2,
		GameServerIp:"192.168.0.166",
		GameServerPort:7400,
		LoginServerIp:"192.168.0.166",
		LoginServerPort:7000,
		Name:"test2",
	})

	db.Delete(Player{})

	return
}

var client *mockclient.MockClient
var cert int

func TestMain(m *testing.M) {

	err := initTestDB()
	if err != nil {
		log.Fatal(err)
	}
	loginChan := make(chan interface{}, 100)

	loginServer := loginserver.New(&loginserver.Option{
		IsTest:true,
		Address:LOGIN_SERVER_ADDRESS,
		DbPath:DB_PATH,
	})
	loginServer.LoginChan = loginChan
	loginServer.Main()

	gameServer := New(&Option{
		IsTest:true,
		Address:GAME_SERVER_ADDRESS,
		DbPath:DB_PATH,
	})
	gameServer.LoginChan = loginChan
	gameServer.Main()

	retCode := m.Run()

	if client != nil {
		client.Close()
	}

	loginServer.Exit()
	gameServer.Exit()
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

func TestLogin(t *testing.T) {
	loginClient, err := mockclient.New(LOGIN_SERVER_ADDRESS)
	defer loginClient.Close()
	if err != nil {
		t.Fatal(err)
	}

	if err := sendAndCheck(loginClient,
		&protocol.Packet{protocol.PacketHeader{0, loginserver.CM_IDPASSWORD, 0, 0, 0}, "pangliang/pwd"},
		&protocol.Packet{protocol.PacketHeader{0, loginserver.SM_PASSOK_SELECTSERVER, 0, 0, 2}, "test1/1/test2/2/"},
	); err != nil {
		t.Fatal(err)
	}

	loginClient.Send(&protocol.Packet{protocol.PacketHeader{0, loginserver.CM_SELECTSERVER, 0, 0, 0}, "test1"})
	resp, err := loginClient.Read()
	if err != nil {
		t.Fatal(fmt.Sprint(err))
	}
	params := resp.Params();
	if (len(params) != 3 || params[0] != "127.0.0.1" || params[1] != "7400") ||
		resp.Header.Protocol != loginserver.SM_SELECTSERVER_OK {
		t.Fatal(fmt.Sprint(resp))
	}
	cert, _ = strconv.Atoi(params[2])

	client, err = mockclient.New(GAME_SERVER_ADDRESS)
	if err != nil {
		t.Fatal(err)
	}

	if err := sendAndCheck(client,
		&protocol.Packet{protocol.PacketHeader{0, CM_QUERYCHR, 0, 0, 0}, fmt.Sprintf("pangliang/%d",cert)},
		&protocol.Packet{protocol.PacketHeader{0, SM_QUERYCHR, 0, 0, 0}, ""},
	); err != nil {
		t.Fatal(err)
	}
}

func TestFailLogin(t *testing.T) {
	newClient, err := mockclient.New(GAME_SERVER_ADDRESS)
	if err != nil {
		t.Fatal(err)
	}

	if err := sendAndCheck(newClient,
		&protocol.Packet{protocol.PacketHeader{0, CM_QUERYCHR, 0, 0, 0}, "pangliang"},
		&protocol.Packet{protocol.PacketHeader{1, SM_QUERYCHR_FAIL, 0, 0, 0}, ""},
	); err != nil {
		t.Fatal(err)
	}

	if err := sendAndCheck(newClient,
		&protocol.Packet{protocol.PacketHeader{0, CM_QUERYCHR, 0, 0, 0}, "pangliang1/1000"},
		&protocol.Packet{protocol.PacketHeader{2, SM_QUERYCHR_FAIL, 0, 0, 0}, ""},
	); err != nil {
		t.Fatal(err)
	}

	if err := sendAndCheck(newClient,
		&protocol.Packet{protocol.PacketHeader{0, CM_QUERYCHR, 0, 0, 0}, "pangliang/1000"},
		&protocol.Packet{protocol.PacketHeader{3, SM_QUERYCHR_FAIL, 0, 0, 0}, ""},
	); err != nil {
		t.Fatal(err)
	}

	if err := sendAndCheck(newClient,
		&protocol.Packet{protocol.PacketHeader{0, CM_NEWCHR, 0, 0, 0}, "pangliang/player1/1/1/1/"},
		&protocol.Packet{protocol.PacketHeader{4, SM_NEWCHR_FAIL, 0, 0, 0}, ""},
	); err != nil {
		t.Fatal(err)
	}
}

func TestCreateDeletePlayer(t *testing.T) {
	if err := sendAndCheck(client,
		&protocol.Packet{protocol.PacketHeader{0, CM_NEWCHR, 0, 0, 0}, "pangliang/player1/3/2/1/"},
		&protocol.Packet{protocol.PacketHeader{0, SM_NEWCHR_SUCCESS, 0, 0, 0}, ""},
	); err != nil {
		t.Fatal(err)
	}

	if err := sendAndCheck(client,
		&protocol.Packet{protocol.PacketHeader{0, CM_NEWCHR, 0, 0, 0}, "pangliang/player2/1/1/2/"},
		&protocol.Packet{protocol.PacketHeader{0, SM_NEWCHR_SUCCESS, 0, 0, 0}, ""},
	); err != nil {
		t.Fatal(err)
	}

	if err := sendAndCheck(client,
		&protocol.Packet{protocol.PacketHeader{0, CM_QUERYCHR, 0, 0, 0}, fmt.Sprintf("pangliang/%d",cert)},
		&protocol.Packet{protocol.PacketHeader{2, SM_QUERYCHR, 0, 0, 0}, "player1/2/3/1/1/player2/1/1/1/2/"},
	); err != nil {
		t.Fatal(err)
	}

	if err := sendAndCheck(client,
		&protocol.Packet{protocol.PacketHeader{0, CM_NEWCHR, 0, 0, 0}, "pangliang/player1/1/1/1/"},
		&protocol.Packet{protocol.PacketHeader{2, SM_NEWCHR_FAIL, 0, 0, 0}, ""},
	); err != nil {
		t.Fatal(err)
	}

	if err := sendAndCheck(client,
		&protocol.Packet{protocol.PacketHeader{0, CM_DELCHR, 0, 0, 0}, "player1"},
		&protocol.Packet{protocol.PacketHeader{0, SM_DELCHR_SUCCESS, 0, 0, 0}, ""},
	); err != nil {
		t.Fatal(err)
	}

	if err := sendAndCheck(client,
		&protocol.Packet{protocol.PacketHeader{0, CM_DELCHR, 0, 0, 0}, "player1"},
		&protocol.Packet{protocol.PacketHeader{2, SM_DELCHR_FAIL, 0, 0, 0}, ""},
	); err != nil {
		t.Fatal(err)
	}
}