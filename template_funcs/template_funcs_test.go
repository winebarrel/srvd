package template_funcs

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func TestTemplateFuncsIpv4ToI(t *testing.T) {
	assert := assert.New(t)
	i := ipv4ToI("1.2.3.4")
	assert.Equal(16909060, i)
}

func TestRotateSRVs(t *testing.T) {
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
