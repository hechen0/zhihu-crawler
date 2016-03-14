package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/mgo.v2"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Question struct {
	Id         int
	Url        string
	Title      string
	AnswerNum  int
	CommentNum int
	Tags       map[string]string
}

var (
	s   = flag.Int("s", 0, "question id start")
	e   = flag.Int("e", 0, "question id end")
	url = flag.String("url", "", "mongodb url")
	db  = flag.String("db", "zhihu", "database")

	regexAnsNum = regexp.MustCompile("^(.*) (.+)$")
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

	session, err := mgo.Dial(*url)
	check(err)
	defer session.Close()

	fmt.Println(*url)
	client := session.DB("zhihu").C("question")

	timer := time.Tick(5 * time.Second)

	var current_num int
	var pre_num int

	pre_num = start

	go func() {
		for range timer {
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

			question := extract_question(i, doc)

			fmt.Println(question)

			err := client.Insert(&question)
			check(err)

			fmt.Printf("%s %s\n", question.Url, question.Title)

		} else {
			title := doc.Find("div.content strong").First().Text()
			title = strings.Trim(title, " \r\n")
			fmt.Println(title)
		}
	}
}

func extract_question(question_id int, doc *goquery.Document) *Question {
	var tmp_str string

	id := question_id
	url := "https://www.zhihu.com/question/" + strconv.Itoa(question_id)

	tmp_str = doc.Find(".zm-item-title").First().Text()
	title := strings.Trim(tmp_str, " \r\n")

	tmp_str, _ = doc.Find("h3#zh-question-answer-num").First().Attr("data-num")
	answer_num, _ := strconv.Atoi(tmp_str)

	tmp_str = doc.Find("div#zh-question-meta-wrap .meta-item").First().Text()
	fields := regexAnsNum.FindStringSubmatch(tmp_str)
	fmt.Printf("%v \n", fields)
	comment_num, _ := strconv.Atoi(strings.Split(tmp_str, " ")[0])

	tags := make(map[string]string)

	return &Question{id, url, title, answer_num, comment_num, tags}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
