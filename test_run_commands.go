package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

var hostname string

type PS struct {
	UID   string
	PID   string
	PPID  string
	C     string
	STIME string
	TTY   string
	TIME  string
	CMD   string
}

// DirSize struct
type DirSize struct {
	FileSystem string
	MBlocks    string
	Used       string
	Avail      string
	Capacity   string
	MountedOn  string
}

//get hostname
func getHostname() string {

	cmdTmp := "hostname"
	cmd := exec.Command("bash", "-c", cmdTmp)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("cmd.Run() failed with %s", err)
	}
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)

	resOutTmp := strings.Split(outStr, "\n")
	for _, row := range resOutTmp {

		if row != "" {
			hostname = row
		}
	}

	return hostname

}

//check filesystem size
func getFilesystem() []DirSize {
	cmd1 := "df -mP"
	cmd := exec.Command("bash", "-c", cmd1)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("cmd.Run() failed with %s", err)
	}
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)

	var ds DirSize
	var dss []DirSize
	resOutTmp := strings.Split(outStr, "\n")
	for i, row := range resOutTmp {
		//skip the header row
		if i == 0 {
			continue
		}
		if row != "" {
			rec := strings.Fields(row)
			//fmt.Println(row)
			//fmt.Println(rec)

			ds.FileSystem = rec[0]
			ds.MBlocks = rec[1]
			ds.Used = rec[2]
			ds.Avail = rec[3]
			ds.Capacity = rec[4]
			ds.MountedOn = rec[5]
		}

		dss = append(dss, ds)

	}

	return dss
}

//get running process
func getProcess() []PS {

	cmdTmp := "ps -ef|grep wcw030"
	cmd := exec.Command("bash", "-c", cmdTmp)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("cmd.Run() failed with %s", err)
	}
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)

	var ps PS
	var pss []PS
	resOutTmp := strings.Split(outStr, "\n")
	for _, row := range resOutTmp {
		if row != "" {
			rec := strings.Fields(row)

			ps.UID = rec[0]
			ps.PID = rec[1]
			ps.PPID = rec[2]
			ps.C = rec[3]
			ps.STIME = rec[4]
			ps.TTY = rec[5]
			ps.TIME = rec[6]
			ps.CMD = strings.Join(rec[7:], " ")
		}

		pss = append(pss, ps)

	}

	return pss
}
func main() {
	host := getHostname()
	fmt.Println("Hostname:")
	fmt.Println(host)

	process := getProcess()
	fmt.Println("Running process:")
	fmt.Println(process)

	fs := getFilesystem()
	fmt.Println("Filesystem size:")
	fmt.Println(fs)

}
