package main

import (
	"fmt"
	"log"
	"time"

	"github.com/okzk/sdnotify"
)

// Worker struct has information on a goroutine which periodically updates the configuration file.
type Worker struct {
	Config     *Config
	StopChan   chan bool
	DoneChan   chan error
	StatusChan chan Status
}

// NewWorker creates Worker structs.
func NewWorker(config *Config, stopChan chan bool, doneChan chan error, statusChan chan Status) (worker *Worker) {
	worker = &Worker{
		Config:     config,
		StopChan:   stopChan,
		DoneChan:   doneChan,
		StatusChan: statusChan,
	}

	return
}

// Run starts a goroutine which periodically updates the configuration file.
func (worker *Worker) Run() {
	defer close(worker.DoneChan)

	dnsCli, err := NewDNSClient(worker.Config)

	if err != nil {
		worker.DoneChan <- fmt.Errorf("DNSClient struct creation failed: %s", err)
		close(worker.StopChan)
		return
	}

	status := Status{}
	tmpl, err := NewTemplate(worker.Config, &status)

	if err != nil {
		worker.DoneChan <- fmt.Errorf("Template struct creation failed: %s", err)
		close(worker.StopChan)
		return
	}

	interval := time.Duration(worker.Config.Interval) * time.Second
	cooldown := time.Duration(worker.Config.Cooldown) * time.Second
	updatedAt := time.Now().Add(-cooldown)

	for {
		srvsByDomain := dnsCli.Dig()
		dnsErr := false
		now := time.Now()

		for domain, srvs := range srvsByDomain {
			if len(srvs) == 0 {
				log.Printf("ERROR: %s SRV record not found", domain)
				dnsErr = true
			}
		}

		if dnsErr {
			status.Ok = false
		} else if updatedAt.Add(cooldown).Before(now) {
			updated := tmpl.Process(srvsByDomain)

			if updated {
				updatedAt = now
				status.LastUpdate = updatedAt
			}
		}

		worker.StatusChan <- status

		if worker.Config.Sdnotify {
			sdnotify.Ready()
			// notify once
			worker.Config.Sdnotify = false
		}

		if worker.Config.Oneshot {
			close(worker.StopChan)
			return
		}

		select {
		case <-worker.StopChan:
			return
		case <-time.After(interval):
			continue
		}
	}
}
