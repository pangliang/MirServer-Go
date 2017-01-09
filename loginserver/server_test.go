package loginserver

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
)

const (
	SERVER_ADDRESS = "127.0.0.1:7000"
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
		"delete from serverinfo",
		"insert into serverinfo values (1,'127.0.0.1',7400,'127.0.0.1',7000,'test1'),(2,'192.168.0.166',7400,'192.168.0.166',7000,'test2')",
	}

	for _,sql := range sqls {
		_, err = db.Exec(sql)
		if err != nil {
			return
		}
	}

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
		DbPath:TEST_DB_PATH,
	}
	loginServer := New(opt)
	loginChan := make(chan interface{}, 100)
	loginServer.LoginChan = loginChan
	loginServer.Main()

	retCode := m.Run()
	loginServer.Exit()
	os.Exit(retCode)
}

func sendAndCheck(client *mockclient.MockClient, request *protocol.Packet, expect *protocol.Packet) (err error){
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
		&protocol.Packet{protocol.PacketHeader{0, CM_ADDNEWUSER, 0, 0, 0}, "11"+string([]byte{0,0,0,0})+"11\x00\x00\x00\x00"},
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

	client.Send(&protocol.Packet{protocol.PacketHeader{0, CM_SELECTSERVER, 0,0,0}, "test1"})
	resp, err := client.Read()
	if err != nil {
		t.Fatal(fmt.Sprint(err))
	}
	if  params := resp.Params(); (len(params) != 3 || params[0] != "127.0.0.1" || params[1] != "7400") ||
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