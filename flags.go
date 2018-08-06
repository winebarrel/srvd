package main

import (
	"flag"
	"fmt"
	"os"
)

var version string

const (
	DefaultConfig = "srvd.toml"
)

type Flags struct {
	Config string
	Dryrun bool
}

func ParseFlag() (flags *Flags) {
	flags = &Flags{}
	var printVersion bool

	flag.StringVar(&flags.Config, "config", DefaultConfig, "config file path")
	flag.BoolVar(&flags.Dryrun, "dryrun", false, "dry run mode")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
	flag.Parse()

	if printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	return
}
