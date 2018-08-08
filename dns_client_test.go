package main

import (
	"regexp"
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/winebarrel/srvd/testutils"
)

func TestDNSClientDig(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Domains:    []string{"_mysql._tcp.winebarrel.jp"},
		ResolvConf: "/etc/resolv.conf",
		Edns0Size:  4096,
	}

	dnsCli, _ := NewDNSClient(config)
	srvsByDomain := dnsCli.Dig()
	assert.Equal(1, len(srvsByDomain))
	srvs := srvsByDomain["_mysql._tcp.winebarrel.jp"]
	assert.Equal(4, len(srvs))

	expects := []string{
		`^_mysql\._tcp\.winebarrel.jp\.` + "\t\\d+\tIN\tSRV\t" + `10 30 80 server4\.winebarrel\.jp\.$`,
		`^_mysql\._tcp\.winebarrel.jp\.` + "\t\\d+\tIN\tSRV\t" + `10 20 80 server3\.winebarrel\.jp\.$`,
		`^_mysql\._tcp\.winebarrel.jp\.` + "\t\\d+\tIN\tSRV\t" + `10 10 80 server1\.winebarrel\.jp\.$`,
		`^_mysql\._tcp\.winebarrel.jp\.` + "\t\\d+\tIN\tSRV\t" + `10 10 80 server2\.winebarrel\.jp\.$`,
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

func TestDNSClientSortSRVs(t *testing.T) {
	assert := assert.New(t)

	srvs := []*dns.SRV{
		&dns.SRV{Priority: 10, Weight: 100, Target: "_http._tcp.bbb.example.com.", Port: 81},
		&dns.SRV{Priority: 10, Weight: 100, Target: "_http._tcp.bbb.example.com.", Port: 80},
		&dns.SRV{Priority: 10, Weight: 100, Target: "_http._tcp.aaa.example.com.", Port: 80},
		&dns.SRV{Priority: 10, Weight: 110, Target: "_http._tcp.aaa.example.com.", Port: 80},
		&dns.SRV{Priority: 20, Weight: 100, Target: "_http._tcp.aaa.example.com.", Port: 80},
	}

	expect := []*dns.SRV{
		&dns.SRV{Priority: 20, Weight: 100, Target: "_http._tcp.aaa.example.com.", Port: 80},
		&dns.SRV{Priority: 10, Weight: 110, Target: "_http._tcp.aaa.example.com.", Port: 80},
		&dns.SRV{Priority: 10, Weight: 100, Target: "_http._tcp.aaa.example.com.", Port: 80},
		&dns.SRV{Priority: 10, Weight: 100, Target: "_http._tcp.bbb.example.com.", Port: 80},
		&dns.SRV{Priority: 10, Weight: 100, Target: "_http._tcp.bbb.example.com.", Port: 81},
	}

	sortSRVs(srvs)
	assert.Equal(expect, srvs)
}

func TestDNSClientDigWithTTL(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Domains:    []string{"_mysql._tcp.example.com"},
		ResolvConf: "/etc/resolv.conf",
	}

	dnsCli, _ := NewDNSClient(config)
	counter := 0

	testutils.PatchMethod(dnsCli.Client, "Exchange", func(guard **monkey.PatchGuard) interface{} {
		return func(_ *dns.Client, _ *dns.Msg, _ string) (r *dns.Msg, _ time.Duration, _ error) {
			answer := []dns.RR{
				&dns.SRV{Priority: 10, Weight: 100, Target: "server1.example.com", Port: 80, Hdr: dns.RR_Header{Ttl: 3}},
				&dns.SRV{Priority: 10, Weight: 100, Target: "server2.example.com", Port: 80, Hdr: dns.RR_Header{Ttl: 3}},
				&dns.SRV{Priority: 10, Weight: 100, Target: "server3.example.com", Port: 80, Hdr: dns.RR_Header{Ttl: 3}},
			}

			if counter == 0 {
				r = &dns.Msg{Answer: answer}
			} else {
				defer (*guard).Unpatch()
				(*guard).Restore()
				r = &dns.Msg{Answer: answer[1:]}
			}

			counter++
			return
		}
	})

	expect := []*dns.SRV{
		&dns.SRV{Priority: 10, Weight: 100, Target: "server1.example.com", Port: 80, Hdr: dns.RR_Header{Ttl: 3}},
		&dns.SRV{Priority: 10, Weight: 100, Target: "server2.example.com", Port: 80, Hdr: dns.RR_Header{Ttl: 3}},
		&dns.SRV{Priority: 10, Weight: 100, Target: "server3.example.com", Port: 80, Hdr: dns.RR_Header{Ttl: 3}},
	}

	srvs1 := dnsCli.Dig()["_mysql._tcp.example.com"]
	assert.Equal(1, counter)
	srvs2 := dnsCli.Dig()["_mysql._tcp.example.com"]
	assert.Equal(1, counter)
	time.Sleep(5 * time.Second)
	srvs3 := dnsCli.Dig()["_mysql._tcp.example.com"]
	assert.Equal(2, counter)
	assert.Equal(expect, srvs1)
	assert.Equal(expect, srvs2)
	assert.Equal(expect[1:], srvs3)
}
