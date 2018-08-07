package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/winebarrel/srvd/testutils"
)

func TestTemplateEvalute(t *testing.T) {
	assert := assert.New(t)
	tmpl := &Template{}

	srvsByDomain := map[string][]*dns.SRV{
		"_mysql._tcp.example.com": []*dns.SRV{&dns.SRV{Target: "server.example.com."}},
	}

	tmplSrc := `{{ $srvs := index .domains "_mysql._tcp.example.com" }}{{ range $srvs }}{{ .Target }}{{ end }}`

	testutils.TempFile(tmplSrc, func(f *os.File) {
		tmpl.Src = f.Name()
		buf, _ := tmpl.evalute(srvsByDomain)
		assert.Equal("server.example.com.", buf.String())
	})
}

func TestTemplateCreateTempDest(t *testing.T) {
	assert := assert.New(t)

	tmpl := &Template{
		DestUID:  os.Getuid(),
		DestGID:  os.Getgid(),
		DestMode: 0644,
	}

	testutils.TempFile("hello", func(f *os.File) {
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

	testutils.TempFile("server0.example.com.", func(dest *os.File) {
		testutils.TempFile("server.example.com.", func(temp *os.File) {
			tmpl.Dest = dest.Name()
			assert.Equal(true, tmpl.isChanged(temp.Name()))
		})
	})
}

func TestTemplateIsChangedFalse(t *testing.T) {
	assert := assert.New(t)
	tmpl := &Template{}

	testutils.TempFile("server.example.com.", func(temp *os.File) {
		testutils.TempFile("server.example.com.", func(dest *os.File) {
			tmpl.Dest = dest.Name()
			assert.Equal(false, tmpl.isChanged(temp.Name()))
		})
	})
}

func TestTemplateIsChangedDestNotExists(t *testing.T) {
	assert := assert.New(t)
	tmpl := &Template{Dest: "not_exists"}

	testutils.TempFile("server.example.com.", func(temp *os.File) {
		assert.Equal(true, tmpl.isChanged(temp.Name()))
	})
}

func TestTemplateUpdate(t *testing.T) {
	assert := assert.New(t)

	tmpl := &Template{
		CheckCmd:  &Command{Cmdline: "true", Timeout: time.Second * time.Duration(3)},
		ReloadCmd: &Command{Cmdline: "true", Timeout: time.Second * time.Duration(3)},
		Config:    &Config{},
	}

	testutils.TempFile("server0.example.com.", func(dest *os.File) {
		testutils.TempFile("server.example.com.", func(temp *os.File) {
			tmpl.Dest = dest.Name()
			err := tmpl.update(temp.Name())
			assert.Equal(nil, err)
			buf, _ := ioutil.ReadFile(dest.Name())
			assert.Equal("server.example.com.", string(buf))
		})
	})
}

func TestTemplateUpdateDryrun(t *testing.T) {
	assert := assert.New(t)

	tmpl := &Template{
		CheckCmd:  &Command{Cmdline: "true", Timeout: time.Second * time.Duration(3)},
		ReloadCmd: &Command{Cmdline: "true", Timeout: time.Second * time.Duration(3)},
		Config:    &Config{Dryrun: true},
	}

	testutils.TempFile("server0.example.com.", func(dest *os.File) {
		testutils.TempFile("server.example.com.", func(temp *os.File) {
			tmpl.Dest = dest.Name()
			err := tmpl.update(temp.Name())
			assert.Equal(nil, err)
			buf, _ := ioutil.ReadFile(dest.Name())
			assert.Equal("server0.example.com.", string(buf))
		})
	})
}

func TestTemplateUpdateCheckFailed(t *testing.T) {
	assert := assert.New(t)

	tmpl := &Template{
		CheckCmd: &Command{Cmdline: "false", Timeout: time.Second * time.Duration(3)},
	}

	testutils.TempFile("server0.example.com.", func(dest *os.File) {
		testutils.TempFile("server.example.com.", func(temp *os.File) {
			tmpl.Dest = dest.Name()
			err := tmpl.update(temp.Name())
			assert.Equal("Check command failed: exit status 1", err.Error())
			buf, _ := ioutil.ReadFile(dest.Name())
			assert.Equal("server0.example.com.", string(buf))
		})
	})
}

func TestTemplateUpdateNoCheck(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		ReloadCmd: "true",
		CheckCmd:  "false",
		Timeout:   3,
		Nocheck:   true,
	}

	tmpl, _ := NewTemplate(config, &Status{})

	testutils.TempFile("server0.example.com.", func(dest *os.File) {
		testutils.TempFile("server.example.com.", func(temp *os.File) {
			tmpl.Dest = dest.Name()
			err := tmpl.update(temp.Name())
			assert.Equal(nil, err)
			buf, _ := ioutil.ReadFile(dest.Name())
			assert.Equal("server.example.com.", string(buf))
		})
	})
}

func TestTemplateUpdateReloadFailed(t *testing.T) {
	assert := assert.New(t)

	tmpl := &Template{
		CheckCmd:  &Command{Cmdline: "true", Timeout: time.Second * time.Duration(3)},
		ReloadCmd: &Command{Cmdline: "false", Timeout: time.Second * time.Duration(3)},
		Config:    &Config{},
	}

	testutils.TempFile("server0.example.com.", func(dest *os.File) {
		testutils.TempFile("server.example.com.", func(temp *os.File) {
			tmpl.Dest = dest.Name()
			err := tmpl.update(temp.Name())
			assert.Equal("Reload command failed: exit status 1", err.Error())
			buf, _ := ioutil.ReadFile(dest.Name())
			assert.Equal("server0.example.com.", string(buf))
		})
	})
}

func TestTemplateUpdateNoReload(t *testing.T) {
	assert := assert.New(t)

	tmpl := &Template{
		CheckCmd:  &Command{Cmdline: "true", Timeout: time.Second * time.Duration(3)},
		ReloadCmd: &Command{Cmdline: "false", Timeout: time.Second * time.Duration(3)},
		Config:    &Config{Noreload: true},
	}

	testutils.TempFile("server0.example.com.", func(dest *os.File) {
		testutils.TempFile("server.example.com.", func(temp *os.File) {
			tmpl.Dest = dest.Name()
			err := tmpl.update(temp.Name())
			assert.Equal(nil, err)
			buf, _ := ioutil.ReadFile(dest.Name())
			assert.Equal("server.example.com.", string(buf))
		})
	})
}

func TestTemplateProcess(t *testing.T) {
	assert := assert.New(t)

	tmpl := &Template{
		CheckCmd:  &Command{Cmdline: "true", Timeout: time.Second * time.Duration(3)},
		ReloadCmd: &Command{Cmdline: "true", Timeout: time.Second * time.Duration(3)},
		DestUID:   os.Getuid(),
		DestGID:   os.Getgid(),
		DestMode:  0644,
		Status:    &Status{},
		Config:    &Config{},
	}

	srvsByDomain := map[string][]*dns.SRV{
		"_mysql._tcp.example.com": []*dns.SRV{&dns.SRV{Target: "server.example.com."}},
	}

	tmplSrc := `{{ $srvs := index .domains "_mysql._tcp.example.com" }}{{ range $srvs }}{{ .Target }}{{ end }}`

	testutils.TempFile("server0.example.com.", func(dest *os.File) {
		testutils.TempFile(tmplSrc, func(src *os.File) {
			tmpl.Dest = dest.Name()
			tmpl.Src = src.Name()
			updated := tmpl.Process(srvsByDomain)
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

	srvsByDomain := map[string][]*dns.SRV{
		"_mysql._tcp.example.com": []*dns.SRV{&dns.SRV{Target: "server.example.com."}},
	}

	tmplSrc := `{{ $srvs := index .domains "_mysql._tcp.example.com" }}{{ range $srvs }}{{ .Target }}{{ end`

	testutils.TempFile(tmplSrc, func(src *os.File) {
		tmpl.Src = src.Name()
		updated := tmpl.Process(srvsByDomain)
		assert.Equal(false, updated)
		assert.Equal(false, tmpl.Status.Ok)
	})
}

func TestTemplateProcessNotChanged(t *testing.T) {
	assert := assert.New(t)

	tmpl := &Template{
		DestUID:  os.Getuid(),
		DestGID:  os.Getgid(),
		DestMode: 0644,
		Status:   &Status{},
	}

	srvsByDomain := map[string][]*dns.SRV{
		"_mysql._tcp.example.com": []*dns.SRV{&dns.SRV{Target: "server.example.com."}},
	}

	tmplSrc := `{{ $srvs := index .domains "_mysql._tcp.example.com" }}{{ range $srvs }}{{ .Target }}{{ end }}`

	testutils.TempFile("server.example.com.", func(dest *os.File) {
		testutils.TempFile(tmplSrc, func(src *os.File) {
			tmpl.Dest = dest.Name()
			tmpl.Src = src.Name()
			updated := tmpl.Process(srvsByDomain)
			assert.Equal(false, updated)
			assert.Equal(true, tmpl.Status.Ok)
			buf, _ := ioutil.ReadFile(dest.Name())
			assert.Equal("server.example.com.", string(buf))
		})
	})
}

func TestTemplateProcessUpdateFailed(t *testing.T) {
	assert := assert.New(t)

	tmpl := &Template{
		CheckCmd: &Command{Cmdline: "false", Timeout: time.Second * time.Duration(3)},
		DestUID:  os.Getuid(),
		DestGID:  os.Getgid(),
		DestMode: 0644,
		Status:   &Status{},
	}

	srvsByDomain := map[string][]*dns.SRV{
		"_mysql._tcp.example.com": []*dns.SRV{&dns.SRV{Target: "server.example.com."}},
	}

	tmplSrc := `{{ $srvs := index .domains "_mysql._tcp.example.com" }}{{ range $srvs }}{{ .Target }}{{ end }}`

	testutils.TempFile("server0.example.com.", func(dest *os.File) {
		testutils.TempFile(tmplSrc, func(src *os.File) {
			tmpl.Dest = dest.Name()
			tmpl.Src = src.Name()
			updated := tmpl.Process(srvsByDomain)
			assert.Equal(false, updated)
			assert.Equal(false, tmpl.Status.Ok)
			buf, _ := ioutil.ReadFile(dest.Name())
			assert.Equal("server0.example.com.", string(buf))
		})
	})
}
