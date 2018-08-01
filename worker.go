package main

import (
	"fmt"
	"time"
)

type Worker struct {
	Config   *Config
	StopChan chan bool
	DoneChan chan error
}

func (worker *Worker) Run() {
	defer close(worker.DoneChan)

	dnsCli, err := NewDnsClient(worker.Config)

	if err != nil {
		worker.DoneChan <- fmt.Errorf("DnsClient struct creation failed: %s", err)
		close(worker.StopChan)
		return
	}

	tmpl, err := NewTemplate(worker.Config)

	if err != nil {
		worker.DoneChan <- fmt.Errorf("Template struct creation failed: %s", err)
		close(worker.StopChan)
		return
	}

	interval := time.Duration(worker.Config.Interval) * time.Second
	cooldown := time.Duration(worker.Config.Cooldown) * time.Second
	updatedAt := time.Now().Add(-cooldown)

	for {
		srvs := dnsCli.Dig()

		if srvs != nil && len(srvs) > 1 {
			now := time.Now()

			if updatedAt.Add(cooldown).Before(now) {
				updated := tmpl.Process(srvs)

				if updated {
					updatedAt = now
				}
			}
		}

		select {
		case <-worker.StopChan:
			return
		case <-time.After(interval):
			continue
		}
	}
}
