package main

import (
	"fmt"
	"log"
	"net"
	"sort"

	"github.com/miekg/dns"
)

type DnsClient struct {
	Domain       string
	ClientConfig *dns.ClientConfig
	Client       *dns.Client
	Message      *dns.Msg
}

func NewDnsClient(config *Config) (dnsCli *DnsClient, err error) {
	dnsCli = &DnsClient{
		Domain:  config.Domain,
		Client:  &dns.Client{},
		Message: &dns.Msg{},
	}

	dnsCli.Message.SetQuestion(dns.Fqdn(config.Domain), dns.TypeSRV)
	dnsCli.Message.RecursionDesired = true
	dnsCli.ClientConfig, err = dns.ClientConfigFromFile(config.ResolvConf)

	return
}

func (dnsCli *DnsClient) Dig() (srvs []*dns.SRV, err error) {
	for _, server := range dnsCli.ClientConfig.Servers {
		hostPort := net.JoinHostPort(server, dnsCli.ClientConfig.Port)
		var r *dns.Msg
		r, _, err = dnsCli.Client.Exchange(dnsCli.Message, hostPort)

		if err != nil {
			log.Println("WARNING: DNS lookup failed: ", err)
		} else if r != nil {
			srvs = make([]*dns.SRV, len(r.Answer))

			for i, a := range r.Answer {
				srvs[i] = a.(*dns.SRV)
			}

			sort.Slice(srvs, func(i, j int) bool {
				return srvs[i].String() < srvs[j].String()
			})

			return
		}
	}

	if err == nil {
		err = fmt.Errorf("DNS record not found")
	}

	return
}
