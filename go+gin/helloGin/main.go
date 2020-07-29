// main.go

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/thda/tds"
)

// server list, need to start with capital letter so that other file in same package can use!
type Server struct {
	Server_name     string
	Server_dns_name string
	Port            int
}

//read servers list
func read_servers() []Server {

	var (
		server_name     string
		server_dns_name string
		port            int
	)

	var servers []Server

	//connection for Sybase to read servers list
	cnxTDSStr := "tds://sscadm:New_DB0@lis-ssc-sd1:32601/SSC_DB?charset=utf8"
	db, err := sql.Open("tds", cnxTDSStr)
	if err != nil {
		log.Println(err)
	}

	defer db.Close()
	// read the server list

	// sql to read server list for sybase
	rows, err := db.Query("select server_name, server_dns_name, port_no from ssc_server where " +
		"environment = 'T' and server_type = 'SYBASE_ASE' and is_active = 'Y' " +
		"and server_name like 'LIS[_]%[_]%' " +
		"and substring(server_name, 5, 3) not in ('ALS', 'PCC', 'PCJ', 'SIT', 'VH5') " +
		"and server_name not like '%ST13%' " +
		"order by server_dns_name, server_name")
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		err = rows.Scan(&server_name, &server_dns_name, &port)
		if err != nil {
			fmt.Println("Failed to scan row", err)
			return nil
		}

		svr := Server{Server_name: server_name, Server_dns_name: server_dns_name, Port: port}
		servers = append(servers, svr)
	}

	defer rows.Close()

	return servers
}

func render(c *gin.Context, data gin.H, templateName string) {

	switch c.Request.Header.Get("Accept") {
	case "application/json":
		// Respond with JSON
		c.JSON(http.StatusOK, data["payload"])
	case "application/xml":
		// Respond with XML
		c.XML(http.StatusOK, data["payload"])
	default:
		// Respond with HTML
		c.HTML(http.StatusOK, templateName, data)
	}

}

func runQuery(query string) []byte {

	var z []byte
	cnxStr := "tds://sscadm:New_DB0@lis-ssc-sd1:32601/SSC_DB?charset=utf8"
	db, err := sql.Open("tds", cnxStr)
	if err != nil {
		log.Println(err)
	}

	rows, err := db.Query(query)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for n := 0; n < 10; n++ {
		fmt.Println(n)
		columnTypes, err := rows.ColumnTypes()

		if err != nil {
			log.Println(err)
		}

		count := len(columnTypes)
		finalRows := []interface{}{}

		for rows.Next() {

			scanArgs := make([]interface{}, count)

			for i, v := range columnTypes {

				switch v.DatabaseTypeName() {
				case "VARCHAR", "TEXT", "UUID", "TIMESTAMP":
					scanArgs[i] = new(sql.NullString)
					break
				case "BOOL":
					scanArgs[i] = new(sql.NullBool)
					break
				case "INT4":
					scanArgs[i] = new(sql.NullInt64)
					break
				default:
					scanArgs[i] = new(sql.NullString)
				}
			}

			err := rows.Scan(scanArgs...)

			if err != nil {
				log.Println(err)
			}

			masterData := map[string]interface{}{}

			for i, v := range columnTypes {

				if z, ok := (scanArgs[i]).(*sql.NullBool); ok {
					masterData[v.Name()] = z.Bool
					continue
				}

				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
					masterData[v.Name()] = z.String
					continue
				}

				if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
					masterData[v.Name()] = z.Int64
					continue
				}

				if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
					masterData[v.Name()] = z.Float64
					continue
				}

				if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
					masterData[v.Name()] = z.Int32
					continue
				}

				masterData[v.Name()] = scanArgs[i]
			}

			finalRows = append(finalRows, masterData)
		}

		z, err = json.Marshal(finalRows)

		// check if have next result set, break if no
		next := rows.NextResultSet()
		fmt.Println("have next? ", next)
		if !next {
			fmt.Printf(string(z))
			break
		}
	}

	return z
}

func showQueryResultJson(c *gin.Context) {

	var r interface{}
	query := c.Param("sql")
	results := runQuery(query)
	err := json.Unmarshal(results, &r)
	if err != nil {
		log.Println(err)
	}
	c.JSON(http.StatusOK, r)
}

func showQueryResult(c *gin.Context) {

	var r interface{}
	query := c.Param("sql")
	results := runQuery(query)
	err := json.Unmarshal(results, &r)
	if err != nil {
		log.Println(err)
	}
	c.HTML(http.StatusOK, "query.html", r)
}

var router *gin.Engine

func main() {
	// Set the router as the default one provided by Gin
	router = gin.Default()

	// Process the templates at the start so that they don't have to be loaded
	// from the disk again. This makes serving HTML pages very fast.
	router.LoadHTMLGlob("templates/*")
	router.Static("/css", "static/css")

	router.GET("/", func(c *gin.Context) {

		servers := read_servers()

		render(c, gin.H{
			"title":   "LIS Test",
			"servers": servers}, "index.html")

	})

	router.GET("/query", func(c *gin.Context) {

		render(c, gin.H{
			"title": "LIS Query",
		}, "query.html")

	})

	router.GET("/query/view", func(c *gin.Context) {

		render(c, gin.H{
			"title": "LIS Query",
		}, "query.html")

	})

	router.GET("/query/json/:sql", showQueryResultJson)

	router.GET("/query/view/:sql", func(c *gin.Context) {

		var r interface{}
		query := c.Param("sql")
		results := runQuery(query)
		err := json.Unmarshal(results, &r)
		if err != nil {
			log.Println(err)
		}

		render(c, gin.H{
			"title":   "LIS Query",
			"results": r,
		}, "query.html")

	})

	// Start serving the application
	router.Run()
}
