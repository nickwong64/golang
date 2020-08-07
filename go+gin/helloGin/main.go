// main.go

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/thda/tds"
)

// Server struct
type Server struct {
	Server_name     string
	Server_dns_name string
	Port_no         string
}

// Raw Result from query
type Result struct {
	Srv     Server
	BResult [][]byte
}

// Final Result from query
type FResult struct {
	Srv     Server
	FResult []interface{}
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

//read servers list
func readServers() (interface{}, error) {

	server := Server{Server_name: "SSC_DB", Server_dns_name: "lis-ssc-sd1", Port_no: "32601"}

	query := "select distinct server_name, server_dns_name, port_no from ssc_server where " +
		"environment = 'T' and server_type = 'SYBASE_ASE' and is_active = 'Y' " +
		"and server_name like 'LIS[_]%[_]%' " +
		"and substring(server_name, 5, 3) not in ('ALS', 'PCC', 'PCJ', 'SIT', 'VH5') " +
		"and server_name not like '%ST13%' " +
		"order by server_dns_name, server_name"

	return runQuery(query, server, "user", "pass", "SSC_DB")

}

// concurrent run query setup
func conrunQuery(wg *sync.WaitGroup, servers <-chan Server, results chan<- FResult, errors chan<- error, query string) {
	for srv := range servers {
		res, err := runQuery(query, srv, "user", "pass", "master")
		if res != nil {
			var r FResult
			r.Srv = srv
			r.FResult = res
			results <- r
		} else if err != nil {
			errors <- err
		}

		wg.Done()
	}

}

// run a single query
func runQuery(query string, srv Server, user string, pass string, dbc string) ([]interface{}, error) {

	// var z []byte
	var zz []interface{}
	cnxStr := fmt.Sprintf("tds://" + user + ":" + pass + "@" + srv.Server_dns_name + ":" + srv.Port_no + "/" + dbc + "?charset=cp850")
	db, err := sql.Open("tds", cnxStr)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		log.Println(err)
		return nil, err
	} else {

		rows, err := db.Query(query)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer rows.Close()

		for n := 0; n < 10; n++ {
			fmt.Println(n)
			columnTypes, err := rows.ColumnTypes()
			columns, err := rows.Columns()
			//fmt.Printf("%#v\n", columns)

			if err != nil {
				log.Println(err)
				return nil, err
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
					return nil, err
				}

				masterData := map[string]interface{}{}

				for i, v := range columns {

					if z, ok := (scanArgs[i]).(*sql.NullBool); ok {
						masterData[v] = z.Bool
						continue
					}

					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						masterData[v] = z.String
						continue
					}

					if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
						masterData[v] = z.Int64
						continue
					}

					if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
						masterData[v] = z.Float64
						continue
					}

					if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
						masterData[v] = z.Int32
						continue
					}

					masterData[v] = scanArgs[i]
				}

				finalRows = append(finalRows, masterData)
				//fmt.Printf("%#v\n", masterData)
			}

			//z, err = json.Marshal(finalRows)

			zz = append(zz, finalRows)

			// check if have next result set, break if no
			next := rows.NextResultSet()
			fmt.Println("have next? ", next)
			if !next {
				//cls
				//fmt.Printf(string(z))

				break
			}
		}

		rows.Close()
		return zz, nil
	}
}

// show servers list as JSON
func showServers(c *gin.Context) {

	servers, err := readServers()
	if err != nil {
		log.Println(err)
	}

	c.JSON(http.StatusOK, servers)
}

var router *gin.Engine

func main() {
	// Set the router as the default one provided by Gin
	router = gin.Default()

	router.Use(cors.Default())

	// Process the templates at the start so that they don't have to be loaded
	// from the disk again. This makes serving HTML pages very fast.
	router.LoadHTMLGlob("templates/*")
	router.Static("/css", "static/css")

	router.GET("/", func(c *gin.Context) {

		servers, err := readServers()
		if err != nil {
			log.Println(err)
		}

		render(c, gin.H{
			"title":   "LIS Test",
			"servers": servers}, "index.html")

	})

	// get server list api
	router.GET("/query/servers", showServers)

	// query page
	router.GET("/query", func(c *gin.Context) {

		render(c, gin.H{
			"title": "LIS Query",
		}, "query.html")

	})

	// submit query api
	router.POST("/query/submit", func(c *gin.Context) {
		fmt.Println("query submitted")
		var servers []Server
		var frs1 []FResult

		//query from post form data
		query := c.PostForm("q")
		// server list from post form data
		serversJson := c.PostForm("s")

		err := json.Unmarshal([]byte(serversJson), &servers)
		if err != nil {
			log.Println(err)
		}
		//fmt.Println("Query : ", query)
		//fmt.Printf("Servers : %#v", servers)

		// below start goroutine to run query
		num := len(servers)
		chan_server := make(chan Server, num)
		chan_Result := make(chan FResult, num)
		errors := make(chan error, 100)

		var wg sync.WaitGroup
		for w := 1; w <= 20; w++ {
			go conrunQuery(&wg, chan_server, chan_Result, errors, query)
		}

		for _, s := range servers {
			chan_server <- s
			wg.Add(1)
		}

		close(chan_server)

		wg.Wait()

		for a := 1; a <= num; a++ {
			select {
			case err := <-errors:
				fmt.Println("have errors:", err.Error())
			case res := <-chan_Result:
				frs1 = append(frs1, res)
				fmt.Println("**********")
			}
		}

		//fmt.Println(frs)
		// sort based on the Server_name
		sort.Slice(frs1[:], func(i, j int) bool {
			return frs1[i].Srv.Server_name < frs1[j].Srv.Server_name
		})
		c.JSON(http.StatusOK, frs1)

	})

	// Start serving the application
	router.Run()
}
