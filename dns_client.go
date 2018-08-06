package main

import (
	"log"
	"net"
	"sort"

	"github.com/miekg/dns"
)

type DnsClient struct {
	ClientConfig *dns.ClientConfig
	Client       *dns.Client
	Messages     map[string]*dns.Msg
}

func NewDnsClient(config *Config) (dnsCli *DnsClient, err error) {
	dnsCli = &DnsClient{
		Client: &dns.Client{},
	}

	dnsCli.Messages = make(map[string]*dns.Msg, len(config.Domains))

	for _, domain := range config.Domains {
		msg := &dns.Msg{}
		msg.SetQuestion(dns.Fqdn(domain), dns.TypeSRV)
		msg.RecursionDesired = true
		dnsCli.Messages[domain] = msg
	}

	dnsCli.ClientConfig, err = dns.ClientConfigFromFile(config.ResolvConf)
	return
}

func (dnsCli *DnsClient) Dig() (srvsByDomain map[string][]*dns.SRV) {
	srvsByDomain = make(map[string][]*dns.SRV, len(dnsCli.Messages))

	for domain, msg := range dnsCli.Messages {
		for _, server := range dnsCli.ClientConfig.Servers {
			hostPort := net.JoinHostPort(server, dnsCli.ClientConfig.Port)
			r, _, err := dnsCli.Client.Exchange(msg, hostPort)

			if err != nil {
				log.Println("WARNING: DNS lookup failed: ", err)
			} else if r != nil {
				srvs := make([]*dns.SRV, len(r.Answer))

				for i, a := range r.Answer {
					srvs[i] = a.(*dns.SRV)
				}

				sort.Slice(srvs, func(i, j int) bool {
					return srvs[i].String() < srvs[j].String()
				})

				srvsByDomain[domain] = srvs
			} else {
				srvsByDomain[domain] = []*dns.SRV{}
			}
		}
	}

	return
}
