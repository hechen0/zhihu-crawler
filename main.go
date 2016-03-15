package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	s   = flag.Int("s", 0, "question id start")
	e   = flag.Int("e", 0, "question id end")
	url = flag.String("url", "", "mongodb url")
	db  = flag.String("db", "zhihu", "database")
)

func main() {

	flag.Parse()

	question_start := *s
	question_end := *e

	start_time := time.Now()

	if question_start == 0 || question_end == 0 || question_start > question_end {
		fmt.Println("question start and end must set correctly")
		return
	}

	out, err := exec.Command("./core.sh").Output()
	check(err)
	total_core, hosts := parse_hosts(out)

	question_total := question_end - question_start

	per_core := question_total / total_core

	var wg sync.WaitGroup

	used_core := 0
	for host := range hosts {
		wg.Add(1)

		go func(host string, core, used_core int) {
			defer wg.Done()

			start := question_start + used_core*per_core
			end := start + core*per_core

			run_machie(start, end, host)

		}(host, hosts[host], used_core)

		used_core += hosts[host]
	}

	wg.Wait()

	elasp := time.Since(start_time)
	fmt.Printf("\n\n from %d to %d, crawl %d; time usage %f min (%f secods) with %d cores \n",
		question_start,
		question_end,
		question_total,
		elasp.Minutes(),
		elasp.Seconds(),
		total_core,
	)
}

func run_machine(start, end int, host string) {
	//-s 20000000 -e 30000000 -url iSource-43-104 -n 43-104

	start_time := time.Now()

	name := strings.Replace(host, ".", "-", -1)

	cmd := exec.Command("ssh", host, "-l", "git", "-p", "10022", "/home/git/zhihu",
		"-s", string(start),
		"-e", string(end),
		"-url", *url,
		"-n", name,
	)

	stdout, err := cmd.StdoutPipe()
	check(err)

	stderr, err := cmd.StderrPipe()
	check(err)

	go io.Copy(os.Stderr, stderr)

	err = cmd.Start()
	check(err)

	io.Copy(os.Stdout, stdout)

	err = cmd.Wait()
	check(err)

	elasp := time.Since(start_time)

	result := fmt.Sprintf("\n\n %s from %d to %d, crawl %d; time usage %f min(%f seconds) \n\n", name, start, end, end-start, elasp.Minutes(), elasp.Seconds())

	fmt.Printf(result)

	log(result)
}

func log(result string) {
	logFile, err := os.OpenFile("machine_log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer logFile.Close()

	if err != nil {
		fmt.Fprintf(logFile, result)
	}
}

func parse_hosts(bytes []byte) (int, map[string]int) {
	var hosts = make(map[string]int)
	total_core := 0

	str := string(bytes[:])
	str = strings.Trim(str, " \r\n")
	lines := strings.Split(str, "\n")
	for i := range lines {
		host := strings.Split(lines[i], " ")
		core, _ := strconv.Atoi(host[1])
		hosts[host[0]] = core
		total_core += core
	}
	return total_core, hosts
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
