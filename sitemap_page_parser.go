package main

import (
	"bytes"
	"net/url"

	sitemap "github.com/oxffaa/gopher-parse-sitemap"
)

type sitemapPageParser struct{}

func newSitemapPageParser() *sitemapPageParser {
	return &sitemapPageParser{}
}

func (f *sitemapPageParser) Parse(u *url.URL, typ string, bs []byte) (page, error) {
	if typ != "application/xml" {
		return nil, nil
	}

	ls := map[string]error{}
	c := func(e interface{ GetLocation() string }) error {
		ls[e.GetLocation()] = nil
		return nil
	}

	err := sitemap.Parse(bytes.NewReader(bs), func(e sitemap.Entry) error {
		return c(e)
	})

	// TODO Detect XML files as sitemaps.
	if err == nil && len(ls) != 0 {
		return newSitemapPage(u, ls), nil
	}

	err = sitemap.ParseIndex(bytes.NewReader(bs), func(e sitemap.IndexEntry) error {
		return c(e)
	})

	// TODO Detect XML files as sitemap indices.
	if err == nil && len(ls) != 0 {
		return newSitemapPage(u, ls), nil
	}

	return nil, nil
}
