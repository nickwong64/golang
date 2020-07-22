// test CRUD operation in sybase table with go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/thda/tds"
)

var ctx context.Context

//insert records
func create(db *sql.DB) {

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	res, err := tx.Exec("insert into dbo.ssc_parameter values ('test2', 'test2', 'test2', 'test2', 'test2')")
	if err != nil {
		log.Fatal(err)
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("ID = %d, affected = %d\n", lastId, rowCnt)

}

//read table
func read(db *sql.DB) {

	rows, err := db.Query("select * from dbo.ssc_parameter where method_name like 'test%'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		fmt.Println("Failed to get columns", err)
		return
	}

	// Result is your slice string.
	rawResult := make([][]byte, len(cols))
	result := make([]string, len(cols))

	dest := make([]interface{}, len(cols)) // A temporary interface{} slice
	for i, _ := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			fmt.Println("Failed to scan row", err)
			return
		}

		for i, raw := range rawResult {
			if raw == nil {
				result[i] = "\\N"
			} else {
				result[i] = string(raw)
			}
		}

		fmt.Printf("%#v\n", result)
	}
}

//update records
func update(db *sql.DB) {

	tx, err := db.Begin()
	res, err := tx.Exec("update dbo.ssc_parameter set module_name = ? where method_name = ? and param_name = ?", "test2-new", "test2", "test2")
	if err != nil {
		log.Fatal(err)
	}

	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("updated = %d\n", rowCnt)

}

//delete records
func delete(db *sql.DB) {

	tx, err := db.Begin()
	res, err := tx.Exec("delete from dbo.ssc_parameter where method_name = ? and param_name = ?", "test2", "test2")
	if err != nil {
		log.Fatal(err)
	}

	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("deleted = %d\n", rowCnt)

}

func numRecordExist(db *sql.DB) {

	var num string
	err := db.QueryRow("select count(*) from dbo.ssc_parameter where method_name like 'test%'").Scan(&num)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(num, "records exists")
}

func main() {
	cnxStr := "tds://sscadm:New_DB0@lis-ssc-sd1:32601/SSC_DB?charset=utf8"
	db, err := sql.Open("tds", cnxStr)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	create(db)
	update(db)
	read(db)
	numRecordExist(db)
	//delete(db)

}
