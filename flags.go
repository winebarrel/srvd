package main

import (
	"flag"
	"fmt"
	"os"
)

var version string

const (
	// DefaultConfig is the default value when the setting file is not specified.
	DefaultConfig = "srvd.toml"
)

// Flags struct has flags passed to srvd.
type Flags struct {
	Config   string
	Dryrun   bool
	Noreload bool
	Nocheck  bool
	Nohttpd  bool
	Oneshot  bool
	Sdnotify bool
}

// ParseFlag parses the flag passed to srvd.
func ParseFlag() (flags *Flags) {
	flags = &Flags{}
	var printVersion bool

	flag.StringVar(&flags.Config, "config", DefaultConfig, "Config file path")
	flag.BoolVar(&flags.Dryrun, "dryrun", false, "Dry run mode")
	flag.BoolVar(&flags.Noreload, "noreload", false, "Skip reloading")
	flag.BoolVar(&flags.Nocheck, "nocheck", false, "Skip checking")
	flag.BoolVar(&flags.Nohttpd, "nohttpd", false, "Stop httpd")
	flag.BoolVar(&flags.Oneshot, "oneshot", false, "Run once")
	flag.BoolVar(&flags.Sdnotify, "sdnotify", false, "Use sd_notify")
	flag.BoolVar(&printVersion, "version", false, "Print version and exit")
	flag.Parse()

	if printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	return
}
