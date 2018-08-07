package main

import (
	"testing"

	"github.com/bouk/monkey"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/winebarrel/srvd/testutils"
)

func TestWorkerUpdated(t *testing.T) {
	assert := assert.New(t)
	workerStopChan := make(chan bool)
	workerDoneChan := make(chan error)
	statusChan := make(chan Status)

	worker := &Worker{
		Config:     &Config{Interval: 60},
		StopChan:   workerStopChan,
		DoneChan:   workerDoneChan,
		StatusChan: statusChan,
	}

	monkey.Patch(NewDNSClient, func(config *Config) (dnsCli *DNSClient, err error) {
		defer monkey.Unpatch(NewDNSClient)
		dnsCli = &DNSClient{}

		testutils.PatchMethod(dnsCli, "Dig", func(guard **monkey.PatchGuard) interface{} {
			return func(_ *DNSClient) (srvsByDomain map[string][]*dns.SRV) {
				defer (*guard).Unpatch()
				(*guard).Restore()

				srvsByDomain = map[string][]*dns.SRV{
					"_mysql._tcp.example.com": []*dns.SRV{&dns.SRV{Target: "server.example.com."}},
				}

				return
			}
		})

		return
	})

	monkey.Patch(NewTemplate, func(config *Config, status *Status) (tmpl *Template, err error) {
		defer monkey.Unpatch(NewTemplate)
		tmpl = &Template{Status: status}

		testutils.PatchMethod(tmpl, "Process", func(guard **monkey.PatchGuard) interface{} {
			return func(tp *Template, _ map[string][]*dns.SRV) (updated bool) {
				defer (*guard).Unpatch()
				(*guard).Restore()
				tp.Status.Ok = true
				updated = true
				return
			}
		})

		return
	})

	var status Status

	go func() {
		status = <-statusChan
		close(workerStopChan)
	}()

	worker.Run()
	assert.Equal(true, status.Ok)
}

func TestWorkerNonUpdated(t *testing.T) {
	assert := assert.New(t)
	workerStopChan := make(chan bool)
	workerDoneChan := make(chan error)
	statusChan := make(chan Status)

	worker := &Worker{
		Config:     &Config{Interval: 60},
		StopChan:   workerStopChan,
		DoneChan:   workerDoneChan,
		StatusChan: statusChan,
	}

	monkey.Patch(NewDNSClient, func(config *Config) (dnsCli *DNSClient, err error) {
		defer monkey.Unpatch(NewDNSClient)
		dnsCli = &DNSClient{}

		testutils.PatchMethod(dnsCli, "Dig", func(guard **monkey.PatchGuard) interface{} {
			return func(_ *DNSClient) (srvsByDomain map[string][]*dns.SRV) {
				defer (*guard).Unpatch()
				(*guard).Restore()

				srvsByDomain = map[string][]*dns.SRV{
					"_mysql._tcp.example.com": []*dns.SRV{&dns.SRV{Target: "server.example.com."}},
				}

				return
			}
		})

		return
	})

	monkey.Patch(NewTemplate, func(config *Config, status *Status) (tmpl *Template, err error) {
		defer monkey.Unpatch(NewTemplate)
		tmpl = &Template{Status: status}

		testutils.PatchMethod(tmpl, "Process", func(guard **monkey.PatchGuard) interface{} {
			return func(tp *Template, _ map[string][]*dns.SRV) (updated bool) {
				defer (*guard).Unpatch()
				(*guard).Restore()
				tp.Status.Ok = false
				updated = false
				return
			}
		})

		return
	})

	var status Status

	go func() {
		status = <-statusChan
		close(workerStopChan)
	}()

	worker.Run()
	assert.Equal(false, status.Ok)
}
