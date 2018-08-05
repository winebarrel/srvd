package main

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDnsClientDig(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Domain:     "_mysql._tcp.winebarrel.jp",
		ResolvConf: "/etc/resolv.conf",
	}

	dnsCli, _ := NewDnsClient(config)
	srvs, err := dnsCli.Dig()
	assert.Equal(nil, err)
	assert.Equal(4, len(srvs))

	expects := []string{
		`^_mysql\._tcp\.winebarrel.jp\.` + "\t\\d+\tIN\tSRV\t" + `10 10 80 server1\.winebarrel\.jp\.$`,
		`^_mysql\._tcp\.winebarrel.jp\.` + "\t\\d+\tIN\tSRV\t" + `10 10 80 server2\.winebarrel\.jp\.$`,
		`^_mysql\._tcp\.winebarrel.jp\.` + "\t\\d+\tIN\tSRV\t" + `10 20 80 server3\.winebarrel\.jp\.$`,
		`^_mysql\._tcp\.winebarrel.jp\.` + "\t\\d+\tIN\tSRV\t" + `10 30 80 server4\.winebarrel\.jp\.$`,
	}

	for i, r := range expects {
		assert.Regexp(regexp.MustCompile(r), srvs[i].String())
	}
}

func TestDnsClientDigNotFound(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Domain:     "_not_exist._tcp.winebarrel.jp",
		ResolvConf: "/etc/resolv.conf",
	}

	dnsCli, _ := NewDnsClient(config)
	srvs, err := dnsCli.Dig()
	assert.Equal(nil, err)
	assert.Equal(0, len(srvs))
}
