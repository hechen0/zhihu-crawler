package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

func main() {
	url := "https://www.zhihu.com/question/36519574"

	headers := map[string]string{
		"user-agent":      "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.116 Safari/537.36",
		"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"cache-control":   "no-cache",
		"accept-language": "accept-language:zh-CN,zh;q=0.8,en-GB;q=0.6,en;q=0.4,fo;q=0.2",
	}
	req, _ := http.NewRequest("GET", url, nil)

	for i := range headers {
		req.Header.Set(i, headers[i])
	}

	fmt.Println("--------- request ----------->")
	for i := range req.Header {
		fmt.Printf("%s: %s\n", i, req.Header[i])
	}

	fmt.Println("--------- response ----------->")
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	for i := range resp.Header {
		fmt.Printf("%s: %s\n", i, resp.Header[i])
	}

	fmt.Println("--------- title ----------->")
	doc, err := goquery.NewDocumentFromResponse(resp)
	check(err)
	title := strings.Trim(doc.Find("title").First().Text(), " \r\n")
	fmt.Println(title)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
