package dao

import (
	"testing"
	"fmt"
	"reflect"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

type User struct {
	Id   uint32
	Name string
	Pwd  string
}

const dbFile = "./test.db"

const ddl = `CREATE TABLE USER
(
ID INTEGER NOT NULL,
NAME TEXT DEFAULT '' NOT NULL,
PWD TEXT DEFAULT '' NOT NULL
);
insert into USER (id,name,pwd) values (1,'111','aaa'),(2,'222','bbb');
`

func getDb() *sql.DB {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	return db
}

func TestReflect(t *testing.T) {
	user := User{99, "pangliang", "pwd"}
	st := reflect.TypeOf(user)
	fmt.Printf("typename:%s : %s\n", st.Name(), st)

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		fmt.Printf("field : %v\n", field)
	}
}

func add(s []int32) {
	s = append(s, 100)
}

func addPtr(s *[]int32) {
	*s = append(*s, 100)
}

func TestSlice(t *testing.T) {
	s := make([]int32, 0, 1)
	fmt.Printf("%v\n", s)
	add(s)
	fmt.Printf("%v\n", s)
	addPtr(&s)
	fmt.Printf("%v\n", s)
}

func TestDaoList(t *testing.T) {
	db := getDb()
	defer db.Close()
	_, err := db.Exec(ddl);
	if err != nil {
		fmt.Printf("error:%s\n", err)
	}

	var userList []User
	err = List(&userList, "", db)
	if err != nil {
		fmt.Printf("error:%s\n", err)
	}
	fmt.Printf("%p -> %v\n", &userList, userList)
}

func TestDeleDbFile(t *testing.T) {
	os.Remove(dbFile)
}