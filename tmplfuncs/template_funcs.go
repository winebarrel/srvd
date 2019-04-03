package tmplfuncs

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/gliderlabs/sigil"
	"github.com/miekg/dns"
)

func init() {
	sigil.Register(template.FuncMap{
		"interfaces":  net.Interfaces,
		"ifaddrs":     net.InterfaceAddrs,
		"ifbyname":    net.InterfaceByName,
		"hostname":    os.Hostname,
		"ipv4byif":    ipv4ByInterface,
		"ipv4sbyif":   ipv4sByInterface,
		"ipv4toi":     ipv4ToI,
		"rotatesrvs":  rotateSRVs,
		"fetchsrvs":   fetchSRVs,
		"shufflesrvs": shuffleSRVs,
		"hextoi":      hexToI,
	})
}

func ipv4sByInterface() (ipv4sByIf map[string][]string, err error) {
	ifs, err := net.Interfaces()

	if err != nil {
		return
	}

	ipv4sByIf = make(map[string][]string, len(ifs))

	for _, i := range ifs {
		buf := []string{}
		addrs, e := i.Addrs()

		if e != nil {
			err = e
			return
		}

		for _, a := range addrs {
			addrStr := a.String()

			if strings.Contains(addrStr, ".") {
				buf = append(buf, addrStr)
			}
		}

		if len(buf) > 0 {
			for i, iprange := range buf {
				idx := strings.Index(iprange, "/")

				if idx > -1 {
					buf[i] = iprange[0:idx]
				}
			}

			ipv4sByIf[i.Name] = buf
		}
	}

	return
}

func ipv4ByInterface() (oneIpv4ByIf map[string]string, err error) {
	ipv4sByIf, err := ipv4sByInterface()

	if err != nil {
		return
	}

	oneIpv4ByIf = make(map[string]string, len(ipv4sByIf))

	for ifname, ipv4s := range ipv4sByIf {
		oneIpv4ByIf[ifname] = ipv4s[0]
	}

	return
}

func ipv4ToI(ip string) (i int) {
	ipv4 := net.ParseIP(ip).To4()

	if ipv4 != nil {
		i = int(ipv4[0])<<24 + int(ipv4[1])<<16 + int(ipv4[2])<<8 + int(ipv4[3])
	}

	return
}

func rotateSRVs(ary []*dns.SRV, n int) []*dns.SRV {
	newAry := make([]*dns.SRV, len(ary))
	copy(newAry, ary)

	if len(newAry) > 0 {
		n = n % len(newAry)
		newAry = append(newAry[n:], newAry[0:n]...)
	}

	return newAry
}

func fetchSRVs(srvsByDomain map[string][]*dns.SRV, domain string) (srvs []*dns.SRV, err error) {
	var ok bool
	srvs, ok = srvsByDomain[domain]

	if !ok {
		err = fmt.Errorf(`Key "%s" not found`, domain)
	}

	return
}

func shuffleSRVs(seed int64, ary []*dns.SRV) []*dns.SRV {
	n := len(ary)
	newAry := make([]*dns.SRV, n)
	copy(newAry, ary)

	src := rand.NewSource(seed)
	rnd := rand.New(src)

	for i := n - 1; i >= 0; i-- {
		j := rnd.Intn(i + 1)
		newAry[i], newAry[j] = newAry[j], newAry[i]
	}

	return newAry
}

func hexToI(hex string) (i int64, err error) {
	i, err = strconv.ParseInt(hex, 16, 64)

	if err != nil {
		if numError, ok := err.(*strconv.NumError); ok {
			if numError.Err == strconv.ErrRange {
				err = nil
			}
		}
	}

	return
}
