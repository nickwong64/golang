// demo the use of concurrent handling from a worker pool
// lookup hostname and also ping if it is alive
package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/tatsushid/go-fastping"
)

type mapping struct {
	ipAddr   string
	hostname string
	alive    bool
}

func ping(ip string) bool {
	pingable := false

	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", ip)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		pingable = true
		//fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
	}
	/*
		p.OnIdle = func() {
			//fmt.Println("finish")
		}*/

	err = p.Run()
	if err != nil {
		fmt.Println(err)
	}

	return pingable
}

func convert_ips3(id int, ips <-chan string, host chan<- mapping) {
	for ip := range ips {
		h, err := net.LookupAddr(ip)
		a := ping(ip)
		m := mapping{}
		if err != nil {
			m = mapping{ip, ip, a}
		} else {
			m = mapping{ip, h[0], a}
		}

		host <- m
	}
}

func main() {

	//list of task to handle
	ip_list := []string{"160.88.49.23", "160.81.64.1", "160.89.57.119", "160.87.83.95", "160.89.57.135", "160.87.83.160", "160.89.57.119", "160.87.83.95", "160.89.57.135", "160.87.83.160"}

	num := len(ip_list)
	ips := make(chan string, num)
	host := make(chan mapping, num)

	//number of workers
	for w := 1; w <= 10; w++ {
		go convert_ips3(w, ips, host)
	}

	//send job
	for _, i := range ip_list {
		ips <- i
	}

	//close job
	close(ips)

	//receive result
	for a := 1; a <= num; a++ {
		fmt.Println(<-host)
	}

}
