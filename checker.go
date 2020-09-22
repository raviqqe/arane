package main

import (
	"sync"

	"github.com/fatih/color"
)

type checker struct {
	fetcher       *linkFetcher
	linkValidator *linkValidator
	daemonManager *daemonManager
	results       chan pageResult
	donePages     concurrentStringSet
	onePageOnly   bool
}

func newChecker(f *linkFetcher, v *linkValidator, concurrency int, onePageOnly bool) *checker {
	return &checker{
		f,
		v,
		newDaemonManager(concurrency),
		make(chan pageResult, concurrency),
		newConcurrentStringSet(),
		onePageOnly,
	}
}

func (c *checker) Results() <-chan pageResult {
	return c.results
}

func (c *checker) Check(page *page) {
	c.addPage(page)
	c.daemonManager.Run()

	close(c.results)
}

func (c *checker) checkPage(p *page) {
	us := p.Links()

	sc := make(chan string, len(us))
	ec := make(chan string, len(us))
	w := sync.WaitGroup{}

	for u, err := range us {
		if err != nil {
			ec <- formatLinkError(u, err)
			continue
		}

		w.Add(1)

		go func(u string) {
			defer w.Done()

			status, p, err := c.fetcher.Fetch(u)

			if err == nil {
				sc <- formatLinkSuccess(u, status)
			} else {
				ec <- formatLinkError(u, err)
			}

			if !c.onePageOnly && p != nil && c.linkValidator.Validate(p.URL()) {
				c.addPage(p)
			}
		}(u)
	}

	w.Wait()

	close(sc)
	close(ec)

	c.results <- newPageResult(p.URL().String(), stringChannelToSlice(sc), stringChannelToSlice(ec))
}

func (c *checker) addPage(p *page) {
	if !c.donePages.Add(p.URL().String()) {
		c.daemonManager.Add(func() { c.checkPage(p) })
	}
}

func stringChannelToSlice(sc <-chan string) []string {
	ss := make([]string, 0, len(sc))

	for s := range sc {
		ss = append(ss, s)
	}

	return ss
}

func formatLinkSuccess(u string, s int) string {
	return color.GreenString("%v", s) + "\t" + u
}

func formatLinkError(u string, err error) string {
	return color.RedString(err.Error()) + "\t" + u
}
