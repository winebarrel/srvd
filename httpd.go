package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Httpd struct has information on httpd which returns srvd status.
type Httpd struct {
	Config     *Config
	StatusChan chan Status
	Status     *Status
}

// NewHttpd creates Httpd struct.
func NewHttpd(config *Config, statusChan chan Status) (httpd *Httpd) {
	httpd = &Httpd{
		Config:     config,
		StatusChan: statusChan,
		Status:     &Status{},
	}

	return
}

func (httpd *Httpd) updateStatus() {
	for status := range httpd.StatusChan {
		httpd.Status = &status
	}
}

func (httpd *Httpd) handler(w http.ResponseWriter, r *http.Request) {
	status, _ := json.Marshal(*httpd.Status)
	fmt.Fprintln(w, string(status))
}

// Run executes httpd.
func (httpd *Httpd) Run() {
	go httpd.updateStatus()

	if !httpd.Config.Nohttpd {
		http.HandleFunc("/status", httpd.handler)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpd.Config.StatusPort), nil))
	}
}
