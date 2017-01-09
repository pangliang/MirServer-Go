package gameserver

import (
	"testing"
	"os"
	"github.com/pangliang/MirServer-Go/mockclient"
	"github.com/pangliang/MirServer-Go/protocol"
	"io/ioutil"
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"errors"
	"fmt"
	"github.com/pangliang/MirServer-Go/loginserver"
)

const (
	LOGIN_SERVER_ADDRESS = "127.0.0.1:7000"
	GAME_SERVER_ADDRESS = "127.0.0.1:7400"
	DB_PATH = "g:/go_workspace/src/github.com/pangliang/MirServer-Go/mir2.db"
	TEST_DB_PATH = DB_PATH + ".test"
)

func initTestDB() (err error) {
	os.Remove(TEST_DB_PATH)
	data, err := ioutil.ReadFile(DB_PATH)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(TEST_DB_PATH, data, 0777)
	if err != nil {
		return
	}

	db, err := sql.Open("sqlite3", TEST_DB_PATH)
	if err != nil {
		return
	}

	sqls := []string{
		"delete from user",
		"insert into user values (1,'pangliang','pwd',0)",
		"delete from serverinfo",
		"insert into serverinfo values (1,'127.0.0.1',7400,'127.0.0.1',7000,'test1'),(2,'192.168.0.166',7400,'192.168.0.166',7000,'test2')",
		"delete from player",
	}

	for _, sqlString := range sqls {
		_, err = db.Exec(sqlString)
		if err != nil {
			return
		}
	}
	return
}

var client *mockclient.MockClient

func TestMain(m *testing.M) {

	err := initTestDB()
	if err != nil {
		log.Fatal(err)
	}
	loginChan := make(chan interface{}, 100)

	loginServer := loginserver.New(&loginserver.Option{
		IsTest:true,
		Address:LOGIN_SERVER_ADDRESS,
		DbPath:TEST_DB_PATH,
	})
	loginServer.LoginChan = loginChan
	loginServer.Main()

	gameServer := New(&Option{
		IsTest:true,
		Address:GAME_SERVER_ADDRESS,
		DbPath:TEST_DB_PATH,
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

	client, err = mockclient.New(GAME_SERVER_ADDRESS)
	if err != nil {
		t.Fatal(err)
	}

	if err := sendAndCheck(client,
		&protocol.Packet{protocol.PacketHeader{0, CM_QUERYCHR, 0, 0, 0}, "pangliang/" + params[2]},
		&protocol.Packet{protocol.PacketHeader{0, SM_QUERYCHR, 0, 0, 0}, ""},
	); err != nil {
		t.Fatal(err)
	}
}