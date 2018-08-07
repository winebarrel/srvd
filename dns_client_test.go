package main

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDNSClientDig(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Domains:    []string{"_mysql._tcp.winebarrel.jp"},
		ResolvConf: "/etc/resolv.conf",
	}

	dnsCli, _ := NewDNSClient(config)
	srvsByDomain := dnsCli.Dig()
	assert.Equal(1, len(srvsByDomain))
	srvs := srvsByDomain["_mysql._tcp.winebarrel.jp"]
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

func TestDNSClientDigNotFound(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Domains:    []string{"_not_exist._tcp.winebarrel.jp"},
		ResolvConf: "/etc/resolv.conf",
	}

	dnsCli, _ := NewDNSClient(config)
	srvsByDomain := dnsCli.Dig()
	assert.Equal(1, len(srvsByDomain))
	srvs := srvsByDomain["_not_exist._tcp.winebarrel.jp"]
	assert.Equal(0, len(srvs))
}
