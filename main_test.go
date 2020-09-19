package main

import (
	"encoding/base64"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommand(t *testing.T) {
	for _, ss := range [][]string{
		{"-x", rootURL},
		{"-j", authorizationHeader("me:password"), basicAuthURL},
		{"-e", ".*", erroneousURL},
	} {
		ok, err := command(ss, ioutil.Discard)

		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestCommandErroneousResult(t *testing.T) {
	for _, ss := range [][]string{
		{erroneousURL},
	} {
		ok, err := command(ss, ioutil.Discard)

		assert.False(t, ok)
		assert.Nil(t, err)
	}
}

func TestCommandError(t *testing.T) {
	for _, ss := range [][]string{
		{":"},
		{"-t", "foo", rootURL},
		{"-j", authorizationHeader("you:password"), basicAuthURL},
	} {
		_, err := command(ss, ioutil.Discard)

		assert.NotNil(t, err)
	}
}

func authorizationHeader(s string) string {
	return "Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte(s))
}
