package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/winebarrel/srvd/testutils"
)

func TestLoadConfig(t *testing.T) {
	assert := assert.New(t)
	flags := &Flags{}

	conf := `
src = "src"
dest = "dest"
domains = ["_http._tcp.example.com"]
reload_cmd = "service reload nginx"
interval = 1
timeout = 2
`

	testutils.TempFile(conf, func(f *os.File) {
		flags.Config = f.Name()
		config, err := LoadConfig(flags)
		assert.Equal(nil, err)
		assert.Equal("src", config.Src)
		assert.Equal("dest", config.Dest)
		assert.Equal([]string{"_http._tcp.example.com"}, config.Domains)
		assert.Equal("/etc/resolv.conf", config.ResolvConf)
		assert.Equal("service reload nginx", config.ReloadCmd)
		assert.Equal("", config.CheckCmd)
		assert.Equal(1, config.Interval)
		assert.Equal(2, config.Timeout)
		assert.Equal(0, config.Cooldown)
		assert.Equal(8080, config.StatusPort)
		assert.Equal(false, config.Dryrun)
		assert.Equal(false, config.Sdnotify)
		assert.Equal(false, config.DisableRollbackOnReloadFailure)
	})
}

func TestLoadConfigWithOptionalConfig(t *testing.T) {
	assert := assert.New(t)
	flags := &Flags{Dryrun: true}

	conf := `
src = "src"
dest = "dest"
domains = ["_http._tcp.example.com"]
resolv_conf = "resolv.conf"
reload_cmd = "service reload nginx"
check_cmd = "service configtest nginx"
interval = 1
timeout = 2
cooldown = 60
status_port = 8081
sdnotify = true
disable_rollback_on_reload_failure = true
`

	testutils.TempFile(conf, func(f *os.File) {
		flags.Config = f.Name()
		config, err := LoadConfig(flags)
		assert.Equal(nil, err)
		assert.Equal("src", config.Src)
		assert.Equal("dest", config.Dest)
		assert.Equal([]string{"_http._tcp.example.com"}, config.Domains)
		assert.Equal("resolv.conf", config.ResolvConf)
		assert.Equal("service reload nginx", config.ReloadCmd)
		assert.Equal("service configtest nginx", config.CheckCmd)
		assert.Equal(1, config.Interval)
		assert.Equal(2, config.Timeout)
		assert.Equal(60, config.Cooldown)
		assert.Equal(8081, config.StatusPort)
		assert.Equal(true, config.Dryrun)
		assert.Equal(true, config.Sdnotify)
		assert.Equal(true, config.DisableRollbackOnReloadFailure)
	})
}

func TestLoadConfigWithoutSrc(t *testing.T) {
	assert := assert.New(t)
	flags := &Flags{}

	conf := `
#src = "src"
dest = "dest"
domains = ["_http._tcp.example.com"]
reload_cmd = "service reload nginx"
interval = 1
timeout = 2
`

	testutils.TempFile(conf, func(f *os.File) {
		flags.Config = f.Name()
		_, err := LoadConfig(flags)
		assert.Equal("src is required", err.Error())
	})
}

func TestLoadConfigWithoutDest(t *testing.T) {
	assert := assert.New(t)
	flags := &Flags{}

	conf := `
src = "src"
#dest = "dest"
domains = ["_http._tcp.example.com"]
reload_cmd = "service reload nginx"
interval = 1
timeout = 2
`

	testutils.TempFile(conf, func(f *os.File) {
		flags.Config = f.Name()
		_, err := LoadConfig(flags)
		assert.Equal("dest is required", err.Error())
	})
}

func TestLoadSrcEqualDest(t *testing.T) {
	assert := assert.New(t)
	flags := &Flags{}

	conf := `
src = "src"
dest = "src"
domains = ["_http._tcp.example.com"]
reload_cmd = "service reload nginx"
interval = 1
timeout = 2
`

	testutils.TempFile(conf, func(f *os.File) {
		flags.Config = f.Name()
		_, err := LoadConfig(flags)
		assert.Equal("src is the same as dest", err.Error())
	})
}

func TestLoadConfigWithoutDomain(t *testing.T) {
	assert := assert.New(t)
	flags := &Flags{}

	conf := `
src = "src"
dest = "dest"
#domaisn =[ "_http._tcp.example.com"]
reload_cmd = "service reload nginx"
interval = 1
timeout = 2
`

	testutils.TempFile(conf, func(f *os.File) {
		flags.Config = f.Name()
		_, err := LoadConfig(flags)
		assert.Equal("domains is required", err.Error())
	})
}

func TestLoadConfigWithoutReoadCmd(t *testing.T) {
	assert := assert.New(t)
	flags := &Flags{}

	conf := `
src = "src"
dest = "dest"
domains = ["_http._tcp.example.com"]
#reload_cmd = "service reload nginx"
interval = 1
timeout = 2
`

	testutils.TempFile(conf, func(f *os.File) {
		flags.Config = f.Name()
		_, err := LoadConfig(flags)
		assert.Equal("reload_cmd is required", err.Error())
	})
}

func TestLoadConfigWithoutInterval(t *testing.T) {
	assert := assert.New(t)
	flags := &Flags{}

	conf := `
src = "src"
dest = "dest"
domains = ["_http._tcp.example.com"]
reload_cmd = "service reload nginx"
#interval = 1
timeout = 2
`

	testutils.TempFile(conf, func(f *os.File) {
		flags.Config = f.Name()
		_, err := LoadConfig(flags)
		assert.Equal("interval mult be '>= 1'", err.Error())
	})
}

func TestLoadConfigWithoutTimeout(t *testing.T) {
	assert := assert.New(t)
	flags := &Flags{}

	conf := `
src = "src"
dest = "dest"
domains = ["_http._tcp.example.com"]
reload_cmd = "service reload nginx"
interval = 1
#timeout = 2
`

	testutils.TempFile(conf, func(f *os.File) {
		flags.Config = f.Name()
		_, err := LoadConfig(flags)
		assert.Equal("timeout mult be '>= 1'", err.Error())
	})
}

func TestLoadConfigWithInvalidStatusPort(t *testing.T) {
	assert := assert.New(t)
	flags := &Flags{}

	conf := `
src = "src"
dest = "dest"
domains = ["_http._tcp.example.com"]
reload_cmd = "service reload nginx"
interval = 1
timeout = 2
status_port = -1
`

	testutils.TempFile(conf, func(f *os.File) {
		flags.Config = f.Name()
		_, err := LoadConfig(flags)
		assert.Equal("status_port mult be '>= 0' && '<= 65535'", err.Error())
	})
}
