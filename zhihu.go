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
	Tags       []int

	UpdatedAt int64
}

type Tag struct {
	Id          int
	Url         string
	Title       string
	Description string

	UpdatedAt int64
}

type Answer struct {
	Id           int
	QuestionId   int
	AnswerId     int
	Url          string
	Vote         int
	Author       string
	AnswerLength int
	CommentNum   int

	CreatedAt int64
	UpdatedAt int64
}

type Selector struct {
	Id int
}

var (
	s   = flag.Int("s", 0, "question id start")
	e   = flag.Int("e", 0, "question id end")
	url = flag.String("url", "", "mongodb url")
	db  = flag.String("db", "zhihu", "database")
	c   = flag.Int("c", runtime.NumCPU(), "worker num")
	n   = flag.String("n", "", "machine name")

	regexCommentNum = regexp.MustCompile("^([0-9]{1}[0-9]+)(\\s+)(.+)$")
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

	if *n == "" {
		fmt.Println("machine name nust be set")
		return
	}

	question_count := question_end - question_start

	worker := *c

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

	elasp := time.Since(start_time)
	fmt.Printf("\n\n from %d to %d, crawl %d ; time usage %f min (%f seconds) \n", question_start, question_end, question_end-question_start, elasp.Minutes(), elasp.Seconds())
}

func run_worker(start, end, crawler int) {

	session, err := mgo.Dial(*url)
	check(err)
	defer session.Close()

	question_collection := session.DB("zhihu").C("question")
	tag_collection := session.DB("zhihu").C("tag")
	answer_collection := session.DB("zhihu").C("answer")

	timer_period := 5
	timer := time.Tick(5 * time.Second)

	var current_num int
	var pre_num int

	pre_num = start

	go func() {
		for range timer {
			t := time.Now().Format("01-31 15:04:15")
			fmt.Printf("%s, %s-#%d: total: %d, current: %d, period: %d/%ds, rate: %d /s \n", t, *n, crawler, end-start, current_num-start, current_num-pre_num, timer_period, (current_num-pre_num)/timer_period)
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
			_, err := question_collection.Upsert(&Selector{question.Id}, question)
			check(err)

			//extract_tags(i, doc)
			tags := extract_tags(i, doc)

			for i := range tags {
				tag := tags[i]
				_, err := tag_collection.Upsert(&Selector{tag.Id}, tag)
				check(err)
			}

			answers := extract_answers(i, doc)
			//extract_answers(i, doc)

			for i := range answers {
				answer := answers[i]
				_, err := answer_collection.Upsert(&Selector{answer.Id}, answer)
				check(err)
			}

			fmt.Printf("%s %s\n", question.Url, question.Title)

		} else if resp.StatusCode == 429 {
			fmt.Printf("429 returned !!!!! %s \n", url)
		}
	}
}

func extract_question(question_id int, doc *goquery.Document) *Question {
	var tmp_str string

	id := question_id
	url := "https://www.zhihu.com/question/" + strconv.Itoa(question_id)

	tmp_str = doc.Find("#zh-question-title").First().Text()
	title := trim(tmp_str)

	tmp_str, _ = doc.Find("h3#zh-question-answer-num").First().Attr("data-num")
	answer_num, _ := strconv.Atoi(tmp_str)

	tmp_str = doc.Find("div#zh-question-meta-wrap .meta-item").First().Text()
	comment_num := extract_comment_num(tmp_str)

	var tags []int
	doc.Find("a.zm-item-tag").Each(func(i int, s *goquery.Selection) {
		tmp_str, exists := s.Attr("href")
		if exists {
			tag_id, err := strconv.Atoi(strings.Split(tmp_str, "/")[2])
			if err == nil {
				tags = append(tags, tag_id)
			}
		}
	})

	now := time.Now().Unix()

	return &Question{id, url, title, answer_num, comment_num, tags, now}
}

func extract_tags(question_id int, doc *goquery.Document) []*Tag {
	var tags []*Tag

	doc.Find("a.zm-item-tag").Each(func(i int, s *goquery.Selection) {
		tag, exists := extract_tag(question_id, s)
		if exists {
			tags = append(tags, tag)
		}
	})

	return tags
}

func extract_tag(question_id int, s *goquery.Selection) (tag *Tag, exists bool) {
	title := trim(s.Text())
	url, exists := s.Attr("href")
	if exists {
		url = trim(url)
		tag_id, _ := strconv.Atoi(strings.Split(url, "/")[2])
		tag = &Tag{tag_id, "https://www.zhihu.com" + url, title, "", time.Now().Unix()}
	}
	return
}

func extract_answers(question_id int, doc *goquery.Document) []*Answer {
	var answers []*Answer

	doc.Find("div.zm-item-answer").Each(func(i int, s *goquery.Selection) {
		answers = append(answers, extract_answer(question_id, s))
	})

	return answers
}

func extract_answer(question_id int, s *goquery.Selection) (answer *Answer) {
	var tmp_str string

	tmp_str, _ = s.Attr("data-aid")
	id, _ := strconv.Atoi(tmp_str)

	tmp_str, _ = s.Attr("data-atoken")
	answer_id, _ := strconv.Atoi(tmp_str)

	url, _ := s.Find(".answer-date-link").Attr("href")
	tmp_str, _ = s.Find(".zm-item-vote-info").Attr("data-votecount")
	vote, _ := strconv.Atoi(tmp_str)
	author, exists := s.Find(".author-link").Attr("href")

	if exists {
		author = author[8:]
	} else {
		author = "ano"
	}

	answer_length := len(s.Find(".zm-editable-content").Text())

	tmp_str = s.Find(".meta-item.toggle-comment").Text()
	comment_num := extract_comment_num(tmp_str)

	tmp_str, _ = s.Attr("data-created")
	created_at, _ := strconv.ParseInt(tmp_str, 10, 64)

	now := time.Now().Unix()

	answer = &Answer{id, question_id, answer_id, "https://www.zhihu.com" + url, vote, author, answer_length, comment_num, created_at, now}
	return
}

func extract_comment_num(str string) (num int) {
	str = trim(str)
	fields := regexCommentNum.FindStringSubmatch(str)
	if len(fields) < 2 {
		num = 0
	} else {
		num, _ = strconv.Atoi(fields[1])
	}
	return
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func trim(str string) string {
	return strings.Trim(str, " \r\n")
}
