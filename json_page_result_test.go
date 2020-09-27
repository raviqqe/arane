package main

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
)

func TestMarshalJSONPageResult(t *testing.T) {
	bs, err := json.Marshal(newJSONPageResult(
		&pageResult{
			"http://foo.com",
			[]*successLinkResult{
				{"http://foo.com/foo", 200},
			},
			[]*errorLinkResult{
				{"http://foo.com/bar", errors.New("baz")},
			},
		}))
	assert.Nil(t, err)
	cupaloy.SnapshotT(t, bs)
}
