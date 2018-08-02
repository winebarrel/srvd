package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Httpd struct {
	Config     *Config
	StatusChan chan Status
	Status     *Status
}

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

func (httpd *Httpd) Run() {
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status, _ := json.Marshal(*httpd.Status)
		fmt.Fprintln(w, string(status))
	})

	go httpd.updateStatus()

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpd.Config.StatusPort), nil))
}
