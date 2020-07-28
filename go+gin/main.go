// main.go

package main

import (
	"database/sql"
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
	cnxTDSStr := "tds://user:password@lis-ssc-sd1:32601/SSC_DB?charset=utf8"
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

var router *gin.Engine

func main() {
	// Set the router as the default one provided by Gin
	router = gin.Default()

	// Process the templates at the start so that they don't have to be loaded
	// from the disk again. This makes serving HTML pages very fast.
	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {

		servers := read_servers()

		render(c, gin.H{
			"title":   "LIS Test",
			"servers": servers}, "index.html")

	})

	// Start serving the application
	router.Run()
}
