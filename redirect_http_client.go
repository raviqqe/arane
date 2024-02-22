package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type redirectHttpClient struct {
	client              httpClient
	maxRedirections     int
	acceptedStatusCodes statusCodeCollection
}

func newRedirectHttpClient(c httpClient, maxRedirections int, acceptedStatusCodes statusCodeCollection) httpClient {
	return &redirectHttpClient{c, maxRedirections, acceptedStatusCodes}
}

func (c *redirectHttpClient) Get(u *url.URL, header http.Header) (httpResponse, error) {
	if header == nil {
		header = http.Header{}
	}

	cj, err := cookiejar.New(nil)

	if err != nil {
		return nil, err
	}

	i := 0

	for {
		for _, c := range cj.Cookies(u) {
			header.Add("cookie", c.String())
		}

		r, err := c.client.Get(u, header)
		if err != nil {
			return nil, c.formatError(err, i, u)
		}

		code := r.StatusCode()

		if c.acceptedStatusCodes.isInCollection(code) {
			return r, nil
		} else if code >= 300 && code <= 399 {
			i++

			if i > c.maxRedirections {
				return nil, errors.New("too many redirections")
			}

			s := r.Header("Location")

			if len(s) == 0 {
				return nil, errors.New("location header not set")
			}

			u, err = u.Parse(s)

			if err != nil {
				return nil, err
			}

			cj.SetCookies(u, parseCookies(r.Header("set-cookie")))
		} else {
			return nil, c.formatError(fmt.Errorf("%v", code), i, u)
		}
	}
}

func (*redirectHttpClient) formatError(err error, redirections int, u *url.URL) error {
	if redirections == 0 {
		return err
	}

	return fmt.Errorf("%w (following redirect %v)", err, u.String())
}

func parseCookies(s string) []*http.Cookie {
	h := http.Header{}
	h.Add("cookie", s)
	return (&http.Request{Header: h}).Cookies()
}
