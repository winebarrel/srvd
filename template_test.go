package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/miekg/dns"

	"github.com/stretchr/testify/assert"
)

func TestEvalute(t *testing.T) {
	assert := assert.New(t)
	tmpl := &Template{}

	srvs := []*dns.SRV{
		&dns.SRV{Target: "server.example.com."},
	}

	tempFile("{{ range .srvs }}{{ .Target }}{{ end }}", func(f *os.File) {
		tmpl.Src = f.Name()
		buf, _ := tmpl.evalute(srvs)
		assert.Equal("server.example.com.", buf.String())
	})
}

func TestCreateTempDest(t *testing.T) {
	assert := assert.New(t)

	tmpl := &Template{
		DestUid:  os.Getuid(),
		DestGid:  os.Getgid(),
		DestMode: 0644,
	}

	tempFile("hello", func(f *os.File) {
		tmpl.Dest = f.Name()
		buf := bytes.NewBufferString("server.example.com.")
		tempPath, _ := tmpl.createTempDest(buf)
		defer os.Remove(tempPath)
		out, _ := ioutil.ReadFile(tempPath)
		assert.Equal("server.example.com.", string(out))
	})
}

func TestIsChangedTrue(t *testing.T) {
	assert := assert.New(t)
	tmpl := &Template{}

	tempFile("server0.example.com.", func(temp *os.File) {
		tempFile("server.example.com.", func(dest *os.File) {
			tmpl.Dest = dest.Name()
			assert.Equal(true, tmpl.isChanged(temp.Name()))
		})
	})
}

func TestIsChangedFalse(t *testing.T) {
	assert := assert.New(t)
	tmpl := &Template{}

	tempFile("server.example.com.", func(temp *os.File) {
		tempFile("server.example.com.", func(dest *os.File) {
			tmpl.Dest = dest.Name()
			assert.Equal(false, tmpl.isChanged(temp.Name()))
		})
	})
}

func TestIsChangedDestNotExists(t *testing.T) {
	assert := assert.New(t)
	tmpl := &Template{Dest: "not_exists"}

	tempFile("server.example.com.", func(temp *os.File) {
		assert.Equal(true, tmpl.isChanged(temp.Name()))
	})
}
