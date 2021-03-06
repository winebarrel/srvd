package main

import (
	"log"
	"net"
	"sort"
	"time"

	"github.com/miekg/dns"
)

// SRVCache struct has SRV record and expiration date.
type SRVCache struct {
	SRVs      []*dns.SRV
	ExpiredAt time.Time
}

// DNSClient struct has DNS query information.
type DNSClient struct {
	ClientConfig *dns.ClientConfig
	Client       *dns.Client
	Messages     map[string]*dns.Msg
	Cache        map[string]*SRVCache
}

// NewDNSClient creates DNSClient struct.
func NewDNSClient(config *Config) (dnsCli *DNSClient, err error) {
	dnsCli = &DNSClient{
		Client: &dns.Client{
			Net: config.Net,
		},
		Cache: map[string]*SRVCache{},
	}

	dnsCli.Messages = make(map[string]*dns.Msg, len(config.Domains))

	for _, domain := range config.Domains {
		msg := &dns.Msg{}
		msg.SetQuestion(dns.Fqdn(domain), dns.TypeSRV)
		msg.RecursionDesired = true
		msg.SetEdns0(config.Edns0Size, true)
		dnsCli.Messages[domain] = msg
	}

	dnsCli.ClientConfig, err = dns.ClientConfigFromFile(config.ResolvConf)
	return
}

// sortSRVs sorts SRVS recors order by Priority Asc, Weight Desc, Target Asc, Port Desc.
func sortSRVs(srvs []*dns.SRV) {
	sort.Slice(srvs, func(i, j int) bool {
		if srvs[i].Priority < srvs[j].Priority { // Asc
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
		if cachedEntry, ok := dnsCli.Cache[domain]; ok {
			if cachedEntry.ExpiredAt.Before(time.Now()) {
				delete(dnsCli.Cache, domain)
			} else {
				srvsByDomain[domain] = cachedEntry.SRVs
				continue
			}
		}

		for _, server := range dnsCli.ClientConfig.Servers {
			hostPort := net.JoinHostPort(server, dnsCli.ClientConfig.Port)
			r, _, err := dnsCli.Client.Exchange(msg, hostPort)

			if err != nil {
				log.Println("WARNING: DNS lookup failed: ", err)
			} else if r.Rcode != dns.RcodeSuccess {
				log.Printf("WARNING: DNS Response Code is not NOERROR: RCODE=%d\n", r.Rcode)
			} else if r != nil {
				srvs := make([]*dns.SRV, len(r.Answer))

				for i, a := range r.Answer {
					srvs[i] = a.(*dns.SRV)
				}

				sortSRVs(srvs)
				srvsByDomain[domain] = srvs

				if len(srvs) > 0 {
					ttl := time.Duration(srvs[0].Hdr.Ttl) * time.Second

					dnsCli.Cache[domain] = &SRVCache{
						SRVs:      srvs,
						ExpiredAt: time.Now().Add(ttl),
					}
				}

				break
			}
		}

		if _, ok := srvsByDomain[domain]; !ok {
			srvsByDomain[domain] = []*dns.SRV{}
		}
	}

	return
}
