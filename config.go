package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

const (
	// DefaultStatusPort is the default status_port value.
	DefaultStatusPort = 8080
	// DefaultEdns0Size is the default edns0_size value.
	DefaultEdns0Size = 4096
)

// Config struct has the setting of srvd.
type Config struct {
	Src                            string
	Dest                           string
	Domains                        []string
	ResolvConf                     string `toml:"resolv_conf"`
	ReloadCmd                      string `toml:"reload_cmd"`
	CheckCmd                       string `toml:"check_cmd"`
	Interval                       int
	Timeout                        int
	Cooldown                       int
	StatusPort                     int `toml:"status_port"`
	Dryrun                         bool
	Noreload                       bool
	Nocheck                        bool
	Nohttpd                        bool
	Oneshot                        bool
	Sdnotify                       bool
	DisableRollbackOnReloadFailure bool   `toml:"disable_rollback_on_reload_failure"`
	Edns0Size                      uint16 `toml:"edns0_size"`
}

// LoadConfig creates Config struct from the given flags.
func LoadConfig(flags *Flags) (config *Config, err error) {
	config = &Config{
		Dryrun:   flags.Dryrun,
		Noreload: flags.Noreload,
		Nocheck:  flags.Nocheck,
		Nohttpd:  flags.Nohttpd,
		Oneshot:  flags.Oneshot,
	}

	if _, e := os.Stat(flags.Config); os.IsNotExist(e) {
		err = fmt.Errorf("Config file not found: %s", flags.Config)
		return
	}

	_, err = toml.DecodeFile(flags.Config, config)

	if config.Src == "" {
		err = fmt.Errorf("src is required")
		return
	}

	if config.Dest == "" {
		err = fmt.Errorf("dest is required")
		return
	}

	if config.Src == config.Dest {
		err = fmt.Errorf("src is the same as dest")
		return
	}

	if len(config.Domains) == 0 {
		err = fmt.Errorf("domains is required")
		return
	}

	if config.ResolvConf == "" {
		config.ResolvConf = "/etc/resolv.conf"
	}

	if config.ReloadCmd == "" {
		err = fmt.Errorf("reload_cmd is required")
		return
	}

	if config.Interval < 1 {
		err = fmt.Errorf("interval mult be '>= 1'")
		return
	}

	if config.Timeout < 1 {
		err = fmt.Errorf("timeout mult be '>= 1'")
		return
	}

	if config.StatusPort == 0 {
		config.StatusPort = DefaultStatusPort
	} else if config.StatusPort < 0 || config.StatusPort > 65535 {
		err = fmt.Errorf("status_port mult be '>= 0' && '<= 65535'")
		return
	}

	if config.Edns0Size < 1 {
		config.Edns0Size = DefaultEdns0Size
	}

	return
}
