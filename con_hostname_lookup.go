// demo the use of concurrent handling from a worker pool
package main
import (
	"net"
	"fmt"
)

type mapping struct {
    ipAddr string
    hostname string
}

func convert_ips2(id int, ips <-chan string, host chan<- mapping) {
	for ip := range ips {
		h, err := net.LookupAddr(ip)
		m := mapping{}
		if err != nil {
			m = mapping {ip, ip}
		} else {
			m = mapping {ip, h[0]}
		}

		host <- m
	}
}

func main() {

	//list of task to handle
	ip_list := []string {"160.88.49.23", "160.81.64.1", "160.89.57.119", "160.87.83.95", "160.89.57.135",  "160.87.83.160", "160.89.57.119", "160.87.83.95", "160.89.57.135", "160.87.83.160"}

	num := len(ip_list)
	ips := make(chan string, num)
	host := make(chan mapping, num)

	//number of workers
	for w := 1; w <= 5; w++ {
		go convert_ips2(w, ips, host)
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
