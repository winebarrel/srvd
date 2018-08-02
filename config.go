package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Src        string
	Dest       string
	Domain     string
	ResolvConf string `toml:"resolv_conf"`
	ReloadCmd  string `toml:"reload_cmd"`
	CheckCmd   string `toml:"check_cmd"`
	Interval   int
	Timeout    int
	Cooldown   int
	StatusPort int `toml:"status_port"`
}

func LoadConfig(flags *Flags) (config *Config, err error) {
	config = &Config{}
	_, err = toml.DecodeFile(flags.Config, config)

	if config.Src == "" {
		err = fmt.Errorf("src is required")
		return
	}

	if config.Dest == "" {
		err = fmt.Errorf("dest is required")
		return
	}

	if config.Domain == "" {
		err = fmt.Errorf("domain is required")
		return
	}

	if config.ResolvConf == "" {
		config.ResolvConf = "/etc/resolv.conf"
		return
	}

	if config.ReloadCmd == "" {
		err = fmt.Errorf("reload is required")
		return
	}

	if config.Interval < 1 {
		err = fmt.Errorf("interval mult be '>= 1'")
		return
	}

	if config.Timeout < 1 {
		err = fmt.Errorf("interval mult be '>= 1'")
		return
	}

	if config.StatusPort == 0 {
		config.StatusPort = 8080
	} else if config.StatusPort < 0 || config.StatusPort > 65535 {
		err = fmt.Errorf("status_port mult be '>= 0' && '<= 65535'")
		return
	}

	return
}
