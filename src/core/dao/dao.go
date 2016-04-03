package dao

import (
	"database/sql"
	"reflect"
	"strings"
	"fmt"
)

func List(v interface{}, where string, db sql.DB) error {


	listPtr := reflect.Indirect(reflect.ValueOf(v))
	daoType := listPtr.Type().Elem()
	listValue := reflect.MakeSlice(listPtr.Type(),0,1)

	tableName := strings.ToUpper(daoType.Name())
	sql := "select * from " + tableName
	if where != "" {
		sql += "where " + where
	}
	fmt.Printf("sql:%s\n", sql)
	rows, err := db.Query(sql)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		fieldsSlice := make([]interface{}, 0, 1)
		for i := 0; i < daoType.NumField(); i++ {
			fieldValue := reflect.New(daoType.Field(i).Type)
			fieldsSlice = append(fieldsSlice, fieldValue.Interface())
		}
		err = rows.Scan(fieldsSlice...)
		if err != nil {
			return err
		}

		obj := reflect.New(daoType).Elem()
		for i := 0; i < daoType.NumField(); i++ {
			obj.Field(i).Set(reflect.Indirect(reflect.ValueOf(fieldsSlice[i])))
		}
		fmt.Println("obj", obj)
		listValue = reflect.Append(listValue, obj)
	}
	listPtr.Set(listValue)
	return nil
}