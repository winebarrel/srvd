package main

import (
	"time"
)

// Status struct has the status of srvd.
type Status struct {
	LastUpdate time.Time
	Ok         bool
}
