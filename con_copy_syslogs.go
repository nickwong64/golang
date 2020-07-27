// test read syslogs from sybase and insert into clickhouse
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

// syslog struct to insert
type syslog struct {
	server_name  string
	DBName       string
	loginame     string
	SPID         int
	starttime    string
	starttime_ms int
	name         string
	//log_datetime string
}

// server list to check syslog
type server struct {
	server_name     string
	server_dns_name string
	port            int
}

//a function to insert record into clickhouse
func insert_clickhouse(db *sql.DB, sls []syslog) {

	//fmt.Println(slice[i:j]) // Process the batch.
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
	}

	stmt, err := tx.Prepare("insert into syslogshold  values (?, ?, ?, ?, ?, ?, ?, now())")
	if err != nil {
		log.Println(err)
	}

	for _, sl := range sls {
		// change the insert SQL here
		_, err = stmt.Exec(sl.server_name, sl.DBName, sl.loginame, sl.SPID, sl.starttime, sl.starttime_ms, sl.name)
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
func read_from_sybase(wg *sync.WaitGroup, servers <-chan server, syslogs chan<- []syslog, errors chan<- error, dbCH *sql.DB) { //dbCH is clickhouse dbTDS is Sybase

	for srv := range servers {
		var (
			servername       string
			loginame         string
			db_name          string
			spid             int
			loggedindatetime time.Time
			hostname         string
			ipaddr           string
			hostprocess      string
			log_datetime     time.Time
		)

		var sls []syslog

		cnxTDSStr := fmt.Sprintf("tds://user:password@" + srv.server_dns_name + ":" + strconv.Itoa(srv.port) + "/master?charset=cp850")
		fmt.Printf("%#v\n", srv)
		dbTDS, err := sql.Open("tds", cnxTDSStr)
		if err != nil {
			log.Println(err)
		}

		//ping if db alive, channel to errors or to syslogs
		err = dbTDS.Ping()
		if err != nil {
			log.Println(err)
			errors <- err
		} else {

			//change the SQL to check syslogs here
			rows, err1 := dbTDS.Query("select distinct @@servername 'server', l.name loginame, db_name(p.dbid) DB, p.spid, " +
				"p. loggedindatetime, coalesce(p.hostname, '') 'hostname', p.ipaddr, p.hostprocess, getdate() log_datetime " +
				"from master..sysprocesses p, master..syslogins l where p.suid = l.suid")
			if err1 != nil {
				//log.Println(err1)
				errors <- err1
			}

			for rows.Next() {
				err2 := rows.Scan(&servername, &loginame, &db_name, &spid, &loggedindatetime, &hostname, &ipaddr, &hostprocess, &log_datetime)
				if err2 != nil {
					log.Println(err2)
					continue
				}

				sl := syslog{server_name: servername, DBName: db_name, loginame: loginame, SPID: spid, starttime: log_datetime.Format("2006-01-02 15:04:05"), starttime_ms: 500, name: hostname}
				sls = append(sls, sl)

			}

			rows.Close()
			syslogs <- sls
		}

		wg.Done()

	}

}

func main() {

	start := time.Now()
	// list of server
	var servers []server
	// list of syslog
	var syslogs []syslog

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

	num := len(servers)
	fmt.Println("number of servers ", strconv.Itoa(num))
	// make channels for server, syslog and errors
	chan_server := make(chan server, num)
	chan_syslog := make(chan []syslog, num)
	errors := make(chan error, 100)

	// setup worker pool to read from sybase
	var wg sync.WaitGroup
	for w := 1; w <= 20; w++ {
		go read_from_sybase(&wg, chan_server, chan_syslog, errors, dbCH)
	}

	// setup channel servers
	for _, s := range servers {
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
		case sl := <-chan_syslog:
			syslogs = append(syslogs, sl...)
		}

	}

	// print out result from syslogs
	fmt.Println("")
	fmt.Println("Print syslogs:")
	for i, sl := range syslogs {
		fmt.Println(i, ".")
		fmt.Printf("%#v\n", sl)
	}

	// insert into clickhouse
	insert_clickhouse(dbCH, syslogs)

	elapsed := time.Since(start)
	log.Println("Time used: %s", elapsed)

}
