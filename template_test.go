package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/miekg/dns"

	"github.com/stretchr/testify/assert"
)

func TestTemplateEvalute(t *testing.T) {
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

func TestTemplateCreateTempDest(t *testing.T) {
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

func TestTemplateIsChangedTrue(t *testing.T) {
	assert := assert.New(t)
	tmpl := &Template{}

	tempFile("server0.example.com.", func(dest *os.File) {
		tempFile("server.example.com.", func(temp *os.File) {
			tmpl.Dest = dest.Name()
			assert.Equal(true, tmpl.isChanged(temp.Name()))
		})
	})
}

func TestTemplateIsChangedFalse(t *testing.T) {
	assert := assert.New(t)
	tmpl := &Template{}

	tempFile("server.example.com.", func(temp *os.File) {
		tempFile("server.example.com.", func(dest *os.File) {
			tmpl.Dest = dest.Name()
			assert.Equal(false, tmpl.isChanged(temp.Name()))
		})
	})
}

func TestTemplateIsChangedDestNotExists(t *testing.T) {
	assert := assert.New(t)
	tmpl := &Template{Dest: "not_exists"}

	tempFile("server.example.com.", func(temp *os.File) {
		assert.Equal(true, tmpl.isChanged(temp.Name()))
	})
}

func TestTemplateUpdate(t *testing.T) {
	assert := assert.New(t)

	tmpl := &Template{
		CheckCmd:  &Command{Cmdline: "true", Timeout: time.Second * time.Duration(3)},
		ReloadCmd: &Command{Cmdline: "true", Timeout: time.Second * time.Duration(3)},
	}

	tempFile("server0.example.com.", func(dest *os.File) {
		tempFile("server.example.com.", func(temp *os.File) {
			tmpl.Dest = dest.Name()
			err := tmpl.update(temp.Name())
			assert.Equal(nil, err)
			buf, _ := ioutil.ReadFile(dest.Name())
			assert.Equal("server.example.com.", string(buf))
		})
	})
}

func TestTemplateUpdateCheckFailed(t *testing.T) {
	assert := assert.New(t)

	tmpl := &Template{
		CheckCmd: &Command{Cmdline: "false", Timeout: time.Second * time.Duration(3)},
	}

	tempFile("server0.example.com.", func(dest *os.File) {
		tempFile("server.example.com.", func(temp *os.File) {
			tmpl.Dest = dest.Name()
			err := tmpl.update(temp.Name())
			assert.Equal("Check command failed: exit status 1", err.Error())
			buf, _ := ioutil.ReadFile(dest.Name())
			assert.Equal("server0.example.com.", string(buf))
		})
	})
}

func TestTemplateUpdateReloadFailed(t *testing.T) {
	assert := assert.New(t)

	tmpl := &Template{
		CheckCmd:  &Command{Cmdline: "true", Timeout: time.Second * time.Duration(3)},
		ReloadCmd: &Command{Cmdline: "false", Timeout: time.Second * time.Duration(3)},
	}

	tempFile("server0.example.com.", func(dest *os.File) {
		tempFile("server.example.com.", func(temp *os.File) {
			tmpl.Dest = dest.Name()
			err := tmpl.update(temp.Name())
			assert.Equal("Reload command failed: exit status 1", err.Error())
			buf, _ := ioutil.ReadFile(dest.Name())
			assert.Equal("server0.example.com.", string(buf))
		})
	})
}

func TestTemplateProcess(t *testing.T) {
	assert := assert.New(t)

	tmpl := &Template{
		CheckCmd:  &Command{Cmdline: "true", Timeout: time.Second * time.Duration(3)},
		ReloadCmd: &Command{Cmdline: "true", Timeout: time.Second * time.Duration(3)},
		DestUid:   os.Getuid(),
		DestGid:   os.Getgid(),
		DestMode:  0644,
		Status:    &Status{},
	}

	srvs := []*dns.SRV{
		&dns.SRV{Target: "server.example.com."},
	}

	tempFile("server0.example.com.", func(dest *os.File) {
		tempFile("{{ range .srvs }}{{ .Target }}{{ end }}", func(src *os.File) {
			tmpl.Dest = dest.Name()
			tmpl.Src = src.Name()
			updated := tmpl.Process(srvs)
			assert.Equal(true, updated)
			assert.Equal(true, tmpl.Status.Ok)
			buf, _ := ioutil.ReadFile(dest.Name())
			assert.Equal("server.example.com.", string(buf))
		})
	})
}

func TestTemplateProcessEvaluteFailed(t *testing.T) {
	assert := assert.New(t)

	tmpl := &Template{
		Status: &Status{},
	}

	srvs := []*dns.SRV{
		&dns.SRV{Target: "server.example.com."},
	}

	tempFile("{{ range .srvs }}{{ .Target }}{{ end", func(src *os.File) {
		tmpl.Src = src.Name()
		updated := tmpl.Process(srvs)
		assert.Equal(false, updated)
		assert.Equal(false, tmpl.Status.Ok)
	})
}

func TestTemplateProcessNotChanged(t *testing.T) {
	assert := assert.New(t)

	tmpl := &Template{
		DestUid:  os.Getuid(),
		DestGid:  os.Getgid(),
		DestMode: 0644,
		Status:   &Status{},
	}

	srvs := []*dns.SRV{
		&dns.SRV{Target: "server.example.com."},
	}

	tempFile("server.example.com.", func(dest *os.File) {
		tempFile("{{ range .srvs }}{{ .Target }}{{ end }}", func(src *os.File) {
			tmpl.Dest = dest.Name()
			tmpl.Src = src.Name()
			updated := tmpl.Process(srvs)
			assert.Equal(false, updated)
			assert.Equal(true, tmpl.Status.Ok)
			buf, _ := ioutil.ReadFile(dest.Name())
			assert.Equal("server.example.com.", string(buf))
		})
	})
}
