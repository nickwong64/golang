// test CRUD operation in sybase table with go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mailru/go-clickhouse"
)

var ctx context.Context

//insert records
func create(db *sql.DB) {

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec("insert into syslogshold  values ('test2', 'test2', 'test2', 3, now(), 500, 'test2', now())")
	if err != nil {
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Created.")
	}

}

//read table
func read(db *sql.DB) {

	rows, err := db.Query("select * from syslogshold s order by server_name")
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

func numRecordExist(db *sql.DB) {

	var num string
	err := db.QueryRow("select count(*) from syslogshold").Scan(&num)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(num, "records exists")
}

func main() {
	cnxStr := "http://user:password@hostname:port_no/database"
	db, err := sql.Open("clickhouse", cnxStr)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	create(db)
	read(db)
	numRecordExist(db)

}
