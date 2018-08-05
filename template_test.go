package main

import (
	"os"
	"testing"

	"github.com/miekg/dns"

	"github.com/stretchr/testify/assert"
)

func TestEvalute(t *testing.T) {
	assert := assert.New(t)
	tmpl := &Template{}

	srvs := []*dns.SRV{
		&dns.SRV{Target: "server.example.com"},
	}

	tempFile("{{ range .srvs }}{{ .Target }}{{ end }}", func(f *os.File) {
		tmpl.Src = f.Name()
		buf, _ := tmpl.evalute(srvs)
		assert.Equal("server.example.com", buf.String())
	})
}
