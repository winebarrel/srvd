package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	log.SetFlags(log.LstdFlags)
}

func main() {
	flags := ParseFlag()
	config, err := LoadConfig(flags)

	if err != nil {
		log.Fatalf("Configuration loading failed: %s", err)
	}

	workerStopChan := make(chan bool)
	workerDoneChan := make(chan error)
	statusChan := make(chan Status)

	worker := NewWorker(config, workerStopChan, workerDoneChan, statusChan)
	go worker.Run()

	httpd := NewHttpd(config, statusChan)
	go httpd.Run()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

LOOP:
	for {
		select {
		case s := <-signalChan:
			log.Printf("Caught %s, Exiting", s)
			close(workerStopChan)
			close(statusChan)
		case e := <-workerDoneChan:
			if e != nil {
				log.Fatalf("FATAL: Processing failed: %s", e)
			} else {
				log.Printf("Exited")
				break LOOP
			}
		}
	}
}
