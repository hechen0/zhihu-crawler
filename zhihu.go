package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	s  = flag.Int("s", 0, "question id start")
	e  = flag.Int("e", 0, "question id end")
	db = flag.String("db", "", "mongodb url")
)

func main() {

	flag.Parse()

	question_start := *s
	question_end := *e

	if question_start == 0 || question_end == 0 || question_start > question_end {
		fmt.Println("question start and end must set correctly")
		return
	}

	question_count := question_end - question_start

	worker := runtime.NumCPU()

	if worker > question_count {
		panic(errors.New("question count must bigger than " + strconv.Itoa(worker)))
	}

	per_worker := (question_end - question_start) / worker

	var wg sync.WaitGroup

	wg.Add(worker)

	for i := 0; i < worker; i++ {
		go func(idx int) {
			defer wg.Done()
			start := question_start + per_worker*idx
			end := question_start + per_worker*(idx+1)
			run_worker(start, end, idx)
		}(i)
	}

	wg.Wait()
}

func run_worker(start, end, crawler int) {
	f, _ := os.OpenFile("urls.txt", os.O_APPEND|os.O_WRONLY, 0600)
	defer f.Close()

	c := time.Tick(5 * time.Second)

	var current_num int
	var pre_num int

	pre_num = start

	go func() {
		for range c {
			fmt.Printf("#%d: current_num: %d, count: %d, rate: %d\n", crawler, current_num, current_num-pre_num, (current_num-pre_num)/5)
			pre_num = current_num
		}
	}()

	for i := start; i < end; i++ {
		current_num = i
		url := "https://www.zhihu.com/question/" + strconv.Itoa(i)
		resp, err := http.Get(url)
		if err != nil {
			continue
		}

		doc, err := goquery.NewDocumentFromResponse(resp)
		if err != nil {
			continue
		}

		if resp.StatusCode == 200 {
			title := doc.Find("title").First().Text()
			title = strings.Trim(title, " \r\n")
			fmt.Printf("%s %s\n", url, title)
			f.WriteString(url + "\n")
		} else {
			title := doc.Find("div.content strong").First().Text()
			title = strings.Trim(title, " \r\n")
			fmt.Println(title)
		}
	}
}

func check_err(err error) {
	if err != nil {
		panic(err)
	}
}
