package main

import (
	"fmt"
	"time"
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
		worker.DoneChan <- fmt.Errorf("DnsClient struct creation failed: %s", err)
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
		now := time.Now()

		if updatedAt.Add(cooldown).Before(now) {
			updated := tmpl.Process(srvsByDomain)

			if updated {
				updatedAt = now
				status.LastUpdate = updatedAt
			}
		}

		worker.StatusChan <- status

		select {
		case <-worker.StopChan:
			return
		case <-time.After(interval):
			continue
		}
	}
}
