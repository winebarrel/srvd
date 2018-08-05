package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHttpdNotOK(t *testing.T) {
	assert := assert.New(t)
	httpd := &Httpd{Status: &Status{}}
	mtx := http.NewServeMux()
	mtx.HandleFunc("/status", httpd.handler)
	ts := httptest.NewServer(mtx)
	defer ts.Close()
	res, _ := http.Get(ts.URL + "/status")
	body, code := readResponse(res)
	assert.Equal(200, code)
	assert.Equal(`{"LastUpdate":"0001-01-01T00:00:00Z","Ok":false}`+"\n", body)
}

func TestHttpdOk(t *testing.T) {
	assert := assert.New(t)
	httpd := &Httpd{Status: &Status{LastUpdate: time.Date(2014, time.December, 31, 12, 13, 24, 0, time.UTC), Ok: true}}
	mtx := http.NewServeMux()
	mtx.HandleFunc("/status", httpd.handler)
	ts := httptest.NewServer(mtx)
	defer ts.Close()
	res, _ := http.Get(ts.URL + "/status")
	body, code := readResponse(res)
	assert.Equal(200, code)
	assert.Equal(`{"LastUpdate":"2014-12-31T12:13:24Z","Ok":true}`+"\n", body)
}
