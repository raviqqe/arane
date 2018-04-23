package main

import (
	"bytes"
	"fmt"
	"net/url"
	"sync"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var validSchemes = map[string]struct{}{
	"":      {},
	"http":  {},
	"https": {},
}

type checker struct {
	fetcher   fetcher
	daemons   daemons
	hostname  string
	results   chan result
	donePages concurrentStringSet
}

func newChecker(s string, c int) (checker, error) {
	f := newFetcher(c)
	p, err := f.Fetch(s)

	if err != nil {
		return checker{}, err
	}

	ch := checker{
		f,
		newDaemons(c),
		p.URL().Hostname(),
		make(chan result, c),
		newConcurrentStringSet(),
	}

	ch.daemons.Add(func() { ch.checkPage(*p) })

	return ch, nil
}

func (c checker) Results() <-chan result {
	return c.results
}

func (c checker) Check() {
	c.daemons.Run()

	close(c.results)
}

func (c checker) checkPage(p page) {
	n, err := html.Parse(bytes.NewReader(p.Body()))

	if err != nil {
		c.results <- newResultWithError(p.URL().String(), err)
		return
	}

	ns := scrape.FindAll(n, func(n *html.Node) bool {
		return n.DataAtom == atom.A
	})

	sc, ec := make(chan string, len(ns)), make(chan string, len(ns))
	w := sync.WaitGroup{}

	for _, n := range ns {
		w.Add(1)

		go func(n *html.Node) {
			defer w.Done()

			u, err := url.Parse(scrape.Attr(n, "href"))

			if err != nil {
				ec <- err.Error()
				return
			}

			if _, ok := validSchemes[u.Scheme]; !ok {
				return
			}

			if !u.IsAbs() {
				u = p.URL().ResolveReference(u)
			}

			p, err := c.fetcher.Fetch(u.String())

			if err == nil {
				sc <- u.String()

				if p != nil && !c.donePages.Add(p.URL().String()) && p.URL().Hostname() == c.hostname {
					c.daemons.Add(func() {
						c.checkPage(*p)
					})
				}
			} else {
				ec <- fmt.Sprintf("%v (%v)", u, err)
			}
		}(n)
	}

	w.Wait()

	c.results <- newResult(p.URL().String(), stringChannelToSlice(sc), stringChannelToSlice(ec))
}

func stringChannelToSlice(sc <-chan string) []string {
	ss := make([]string, 0, len(sc))

	for i := 0; i < cap(ss); i++ {
		ss = append(ss, <-sc)
	}

	return ss
}
