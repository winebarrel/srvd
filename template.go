package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"github.com/gliderlabs/sigil"
	"github.com/miekg/dns"
)

type Template struct {
	Src       string
	Dest      string
	DestMode  os.FileMode
	DestUid   int
	DestGid   int
	CheckCmd  *Command
	ReloadCmd *Command
}

func NewTemplate(config *Config) (tmpl *Template, err error) {
	tmpl = &Template{
		Src:       config.Src,
		Dest:      config.Dest,
		DestMode:  0644,
		DestUid:   os.Getuid(),
		DestGid:   os.Getgid(),
		ReloadCmd: NewCommand(config.ReloadCmd, config.Timeout),
	}

	if config.CheckCmd != "" {
		tmpl.CheckCmd = NewCommand(config.CheckCmd, config.Timeout)
	}

	_, err = os.Stat(tmpl.Src)

	if err != nil {
		return
	}

	if s, e := os.Stat(tmpl.Dest); e == nil {
		tmpl.DestMode = s.Mode()
		stat := s.Sys().(*syscall.Stat_t)
		tmpl.DestUid = int(stat.Uid)
		tmpl.DestGid = int(stat.Gid)
	}

	return
}

func (tmpl *Template) evalute(srvs []*dns.SRV) (buf bytes.Buffer, err error) {
	input, err := ioutil.ReadFile(tmpl.Src)

	if err != nil {
		return
	}

	vars := map[string]interface{}{
		"srvs": srvs,
	}
	name := filepath.Base(tmpl.Src)
	buf, err = sigil.Execute(input, vars, name)
	return
}

func (tmpl *Template) createTempDest(buf bytes.Buffer) (tempPath string, err error) {
	destTemp, err := ioutil.TempFile(filepath.Dir(tmpl.Dest), "."+filepath.Base(tmpl.Dest))

	if err != nil {
		return
	}

	defer destTemp.Close()
	tempPath = destTemp.Name()
	os.Chown(tempPath, tmpl.DestUid, tmpl.DestGid)
	os.Chmod(tempPath, tmpl.DestMode)
	_, err = destTemp.Write(buf.Bytes())
	return
}

func (tmpl *Template) isChanged(tempPath string) bool {
	if _, err := os.Stat(tmpl.Dest); os.IsNotExist(err) {
		return true
	}

	destMd5 := Md5(tmpl.Dest)
	tempMd5 := Md5(tempPath)
	return destMd5 != tempMd5
}

func (tmpl *Template) update(tempPath string) (err error) {
	if tmpl.CheckCmd != nil {
		log.Printf("Run '%s' for checking", tmpl.CheckCmd.Cmdline)
		err = tmpl.CheckCmd.Run(tempPath)

		if err != nil {
			err = fmt.Errorf("Check command failed: %s", err)
			return
		}
	}

	err = os.Rename(tempPath, tmpl.Dest)

	if err != nil {
		return
	}

	log.Printf("Run '%s' for reloading", tmpl.ReloadCmd.Cmdline)
	err = tmpl.ReloadCmd.Run(tempPath)

	if err != nil {
		err = fmt.Errorf("Reload command failed: %s", err)
		return
	}

	return
}

func (tmpl *Template) Process(srvs []*dns.SRV) (updated bool) {
	buf, err := tmpl.evalute(srvs)

	if err != nil {
		log.Println("ERROR: Template evaluating failed:", err)
		return
	}

	tempPath, err := tmpl.createTempDest(buf)

	if err != nil {
		log.Println("ERROR: Temporary dest file creation failed:", err)
		return
	}

	defer os.Remove(tempPath)

	if tmpl.isChanged(tempPath) {
		log.Println("The configuration has been changed. Update", tmpl.Dest)
		err = tmpl.update(tempPath)

		if err != nil {
			log.Println("ERROR: Temporary dest file creation failed:", err)
			return
		}

		updated = true
	}

	return
}
