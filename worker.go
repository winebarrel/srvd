package main

import (
	"fmt"
	"log"
	"time"
)

type Worker struct {
	Config     *Config
	StopChan   chan bool
	DoneChan   chan error
	StatusChan chan Status
}

func NewWorker(config *Config, stopChan chan bool, doneChan chan error, statusChan chan Status) (worker *Worker) {
	worker = &Worker{
		Config:     config,
		StopChan:   stopChan,
		DoneChan:   doneChan,
		StatusChan: statusChan,
	}

	return
}

func (worker *Worker) Run() {
	defer close(worker.DoneChan)

	dnsCli, err := NewDnsClient(worker.Config)

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
		srvs, err := dnsCli.Dig()

		if err != nil {
			log.Println("DNS request failed:", err)
			status.Ok = false
		} else if srvs != nil && len(srvs) > 0 {
			now := time.Now()

			if updatedAt.Add(cooldown).Before(now) {
				updated := tmpl.Process(srvs)

				if updated {
					updatedAt = now
					status.LastUpdate = updatedAt
				}
			}
		} else {
			log.Fatalf("Invalid DNS records detected: %v", srvs)
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
