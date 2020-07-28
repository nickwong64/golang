// test read sysprocesses from sybase and insert into clickhouse
package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	_ "github.com/mailru/go-clickhouse"
	_ "github.com/thda/tds"
)

// syspr struct to insert
type syspr struct {
	server           string
	loginame         string
	DB               string
	spid             int
	loggedindatetime string
	hostname         string
	ipaddr           string
	hostprocess      string
	log_datetime     string
}

// server list to check syspr
type server struct {
	server_name     string
	server_dns_name string
	port            int
}

//a function to insert record into clickhouse
func insert_clickhouse(db *sql.DB, sps []syspr) {

	//fmt.Println(slice[i:j]) // Process the batch.
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
	}

	stmt, err := tx.Prepare("insert into sysprocesses values (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Println(err)
	}

	for _, sp := range sps {
		// change the insert SQL here
		_, err = stmt.Exec(sp.server, sp.loginame, sp.DB, sp.spid, sp.loggedindatetime, sp.hostname, sp.ipaddr, sp.hostprocess, sp.log_datetime)
		if err != nil {
			log.Println(err)
		}
	}
	fmt.Println("---")

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	} else {
	}

}

//read servers list
func read_servers(db *sql.DB) []server {

	var (
		server_name     string
		server_dns_name string
		port            int
	)

	var servers []server

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

		svr := server{server_name: server_name, server_dns_name: server_dns_name, port: port}
		servers = append(servers, svr)
	}

	defer rows.Close()

	return servers
}

//read from sybase sysprocesses, concurrent
func read_from_sybase(wg *sync.WaitGroup, servers <-chan server, sysprs chan<- []syspr, errors chan<- error, dbCH *sql.DB) { //dbCH is clickhouse dbTDS is Sybase

	for srv := range servers {
		var (
			server           string
			loginame         string
			DB               string
			spid             int
			loggedindatetime time.Time
			hostname         string
			ipaddr           string
			hostprocess      string
			log_datetime     time.Time
		)

		var sps []syspr

		cnxTDSStr := fmt.Sprintf("tds://user:password@" + srv.server_dns_name + ":" + strconv.Itoa(srv.port) + "/master?charset=cp850")
		fmt.Printf("%#v\n", srv)
		dbTDS, err := sql.Open("tds", cnxTDSStr)
		if err != nil {
			log.Println(err)
		}

		//ping if db alive, channel to errors or to sysprs
		err = dbTDS.Ping()
		if err != nil {
			//log.Println(err)
			errors <- err
		} else {

			//change the SQL to check sysprs here
			rows, err1 := dbTDS.Query("select distinct @@servername 'server', l.name loginame, db_name(p.dbid) DB, p.spid, p. loggedindatetime, " +
				"coalesce(p.hostname, '') 'hostname', p.ipaddr, p.hostprocess, getdate() log_datetime " +
				"from master..sysprocesses p, master..syslogins l " +
				"where p.suid = l.suid " +
				"order by server, loginame")
			if err1 != nil {
				//log.Println(err1)
				//log.Println("can see me?")
				errors <- err1
			} else {

				for rows.Next() {
					err2 := rows.Scan(&server, &loginame, &DB, &spid, &loggedindatetime, &hostname, &ipaddr, &hostprocess, &log_datetime)
					if err2 != nil {
						log.Println(err2)
					} else {

						sp := syspr{server: server, loginame: loginame, DB: DB, spid: spid, loggedindatetime: loggedindatetime.Format("2006-01-02 15:04:05"), hostname: hostname, ipaddr: ipaddr, hostprocess: hostprocess, log_datetime: log_datetime.Format("2006-01-02 15:04:05")}
						sps = append(sps, sp)
					}

				}

				rows.Close()
				sysprs <- sps
			}
		}

		wg.Done()

	}

}

func main() {

	start := time.Now()
	// list of server
	var servers []server
	// list of syspr
	var sysprs []syspr

	//connection for Sybase to read servers list
	cnxTDSStr := "tds://user:password@lis-ssc-sd1:32601/SSC_DB?charset=utf8"
	dbServers, err := sql.Open("tds", cnxTDSStr)
	if err != nil {
		log.Println(err)
	}

	//connection for clickhouse
	cnxCHStr := "http://user:password@lisvmc2c:8123/MON_DB"
	dbCH, err := sql.Open("clickhouse", cnxCHStr)
	if err != nil {
		log.Println(err)
	}

	defer dbCH.Close()
	defer dbServers.Close()
	// read the server list
	servers = read_servers(dbServers)

	num := len(servers[:])
	fmt.Println("number of servers ", strconv.Itoa(num))
	// make channels for server, syspr and errors
	chan_server := make(chan server, num)
	chan_syspr := make(chan []syspr, num)
	errors := make(chan error, 100)

	// setup worker pool to read from sybase
	var wg sync.WaitGroup
	for w := 1; w <= 20; w++ {
		go read_from_sybase(&wg, chan_server, chan_syspr, errors, dbCH)
	}

	// setup channel servers
	for _, s := range servers[:] {
		chan_server <- s
		wg.Add(1)
	}

	close(chan_server)

	wg.Wait()

	// setup channel to receive
	for a := 1; a <= num; a++ {

		select {
		case err := <-errors:
			fmt.Println("have error:", err.Error())
		case sp := <-chan_syspr:
			sysprs = append(sysprs, sp...)
		}

	}

	// print out result from sysprs
	fmt.Println("")
	fmt.Println("Print sysprs:")
	for i, sp := range sysprs {
		fmt.Println(i, ".")
		fmt.Printf("%#v\n", sp)
	}

	// insert into clickhouse
	insert_clickhouse(dbCH, sysprs)

	elapsed := time.Since(start)
	log.Println("Time used: %s", elapsed)

}
