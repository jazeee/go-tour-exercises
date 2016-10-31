package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type Result struct {
	body string
	urls []string
	err  error
}
type ResultsCache struct {
	cache map[string]Result
	mux   sync.Mutex
}

var resultsCache ResultsCache = ResultsCache{cache: make(map[string]Result)}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, c chan int) {
	// TODO: Fetch URLs in parallel.
	defer close(c)
	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	resultsCache.mux.Lock()
	_, exists := resultsCache.cache[url]
	if exists {
		resultsCache.mux.Unlock()
		return
	}
	resultsCache.cache[url] = Result{body: body, urls: urls, err: err}
	resultsCache.mux.Unlock()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	awaitChans := make([]chan int, 0, 100)
	for _, u := range urls {
		childChan := make(chan int)
		n := len(awaitChans)
		awaitChans = awaitChans[0 : n+1]
		awaitChans[n] = childChan
		go Crawl(u, depth-1, fetcher, childChan)
	}
	for _,awaitChan := range awaitChans {
		<- awaitChan
	}
	return
}

func main() {
	c := make(chan int)
	go Crawl("http://golang.org/", 4, fetcher, c)
	<-c
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}

