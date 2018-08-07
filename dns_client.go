package main

import (
	"log"
	"net"
	"sort"

	"github.com/miekg/dns"
)

// DNSClient struct has DNS query information.
type DNSClient struct {
	ClientConfig *dns.ClientConfig
	Client       *dns.Client
	Messages     map[string]*dns.Msg
}

// NewDNSClient creates DNSClient struct.
func NewDNSClient(config *Config) (dnsCli *DNSClient, err error) {
	dnsCli = &DNSClient{
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

// sortSRVs sorts SRVS recors order by Priority Desc, Weight Desc, Target Asc, Port Desc
func sortSRVs(srvs []*dns.SRV) {
	sort.Slice(srvs, func(i, j int) bool {
		if srvs[i].Priority > srvs[j].Priority { // Desc
			return true
		} else if srvs[i].Priority == srvs[j].Priority {
			if srvs[i].Weight > srvs[j].Weight { // Desc
				return true
			} else if srvs[i].Weight == srvs[j].Weight {
				if srvs[i].Target < srvs[j].Target { // Asc
					return true
				} else if srvs[i].Target == srvs[j].Target {
					if srvs[i].Port < srvs[j].Port { // Asc
						return true
					}
				}
			}
		}

		return false
	})
}

// Dig queries the SRV record.
func (dnsCli *DNSClient) Dig() (srvsByDomain map[string][]*dns.SRV) {
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

				sortSRVs(srvs)
				srvsByDomain[domain] = srvs
			} else {
				srvsByDomain[domain] = []*dns.SRV{}
			}
		}
	}

	return
}
