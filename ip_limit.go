package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
)

var (
	url = flag.String("url", "", "url to test")
	c   = flag.Int("c", 10, "concurrency")
)

func main() {
	flag.Parse()

	var wg sync.WaitGroup

	for i := 0; i < *c; i++ {
		wg.Add(1)
		go func() {
			resp, err := http.Get(*url)

			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(resp.StatusCode)
			}

			wg.Done()
		}()
	}

	wg.Wait()
	fmt.Println("DONE")
}
