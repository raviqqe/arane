package main

import (
	"bytes"
	"net/url"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type htmlPageParser struct {
	linkFinder linkFinder
}

func newHtmlPageParser(f linkFinder) *htmlPageParser {
	return &htmlPageParser{f}
}

func (p htmlPageParser) Parse(u *url.URL, typ string, body []byte) (page, error) {
	if typ != "text/html" {
		return nil, nil
	}

	n, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	u.Fragment = ""

	frs := map[string]struct{}{}

	scrape.FindAllNested(n, func(n *html.Node) bool {
		for _, a := range []string{"id", "name"} {
			if s := scrape.Attr(n, a); s != "" {
				frs[s] = struct{}{}
			}
		}

		return false
	})

	base := u

	if n, ok := scrape.Find(n, func(n *html.Node) bool {
		return n.DataAtom == atom.Base
	}); ok {
		u, err := url.Parse(scrape.Attr(n, "href"))
		if err != nil {
			return nil, err
		}

		base = base.ResolveReference(u)
	}

	return newHtmlPage(u, frs, p.linkFinder.Find(n, base)), nil
}
