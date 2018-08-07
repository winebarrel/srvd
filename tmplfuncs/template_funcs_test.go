package tmplfuncs

import (
	"net"
	"testing"

	"github.com/bouk/monkey"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/winebarrel/srvd/testutils"
)

func TestTemplateIpv4sByInterface(t *testing.T) {
	monkey.Patch(net.Interfaces, func() (ifs []net.Interface, err error) {
		defer monkey.Unpatch(net.Interfaces)
		i := &net.Interface{Name: "eth0"}
		ifs = []net.Interface{*i}

		testutils.PatchMethod(i, "Addrs", func(guard **monkey.PatchGuard) interface{} {
			return func(_ *net.Interface) (addrs []net.Addr, _ error) {
				defer (*guard).Unpatch()
				(*guard).Restore()

				addrs = []net.Addr{
					&net.IPNet{
						IP:   net.IPv4(192, 168, 0, 1),
						Mask: net.IPv4Mask(255, 255, 255, 0),
					},
				}

				return
			}
		})

		return
	})

	assert := assert.New(t)
	ipv4sByIf, _ := ipv4sByInterface()
	assert.Equal(map[string][]string(map[string][]string{"eth0": []string{"192.168.0.1"}}), ipv4sByIf)
}

func TestTemplateIpv4ByInterface(t *testing.T) {
	monkey.Patch(net.Interfaces, func() (ifs []net.Interface, err error) {
		defer monkey.Unpatch(net.Interfaces)
		i := &net.Interface{Name: "eth0"}
		ifs = []net.Interface{*i}

		testutils.PatchMethod(i, "Addrs", func(guard **monkey.PatchGuard) interface{} {
			return func(_ *net.Interface) (addrs []net.Addr, _ error) {
				defer (*guard).Unpatch()
				(*guard).Restore()

				addrs = []net.Addr{
					&net.IPNet{
						IP:   net.IPv4(192, 168, 0, 1),
						Mask: net.IPv4Mask(255, 255, 255, 0),
					},
				}

				return
			}
		})

		return
	})

	assert := assert.New(t)
	ipv4ByIf, _ := ipv4ByInterface()
	assert.Equal(map[string]string{"eth0": "192.168.0.1"}, ipv4ByIf)
}

func TestTemplateFuncsIpv4ToI(t *testing.T) {
	assert := assert.New(t)
	i := ipv4ToI("1.2.3.4")
	assert.Equal(16909060, i)
}

func TestTemplateFuncsRotateSRVs(t *testing.T) {
	assert := assert.New(t)

	ary := []*dns.SRV{
		&dns.SRV{Target: "1"},
		&dns.SRV{Target: "2"},
		&dns.SRV{Target: "3"},
		&dns.SRV{Target: "4"},
		&dns.SRV{Target: "5"},
	}

	actual := rotateSRVs(ary, 3)
	assert.Equal(5, len(actual))
	assert.Equal(actual[0].Target, "4")
	assert.Equal(actual[1].Target, "5")
	assert.Equal(actual[2].Target, "1")
	assert.Equal(actual[3].Target, "2")
	assert.Equal(actual[4].Target, "3")
}

func TestTemplateFuncsFetchSRVs(t *testing.T) {
	assert := assert.New(t)
	srvsByDomain := map[string][]*dns.SRV{"exist": []*dns.SRV{}}
	srvs, err := fetchSRVs(srvsByDomain, "exist")
	assert.Equal([]*dns.SRV{}, srvs)
	assert.Equal(nil, err)
	_, err = fetchSRVs(srvsByDomain, "not_exist")
	assert.Equal(`Key "not_exist" not found`, err.Error())
}
