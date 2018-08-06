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
	_ "github.com/gliderlabs/sigil/builtin"
	"github.com/miekg/dns"
	_ "github.com/winebarrel/srvd/template_funcs"
)

type Template struct {
	Src       string
	Dest      string
	DestMode  os.FileMode
	DestUid   int
	DestGid   int
	CheckCmd  *Command
	ReloadCmd *Command
	Status    *Status
	Dryrun    bool
}

func NewTemplate(config *Config, status *Status) (tmpl *Template, err error) {
	tmpl = &Template{
		Src:       config.Src,
		Dest:      config.Dest,
		DestMode:  0644,
		DestUid:   os.Getuid(),
		DestGid:   os.Getgid(),
		ReloadCmd: NewCommand(config.ReloadCmd, config.Timeout),
		Status:    status,
		Dryrun:    config.Dryrun,
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

func (tmpl *Template) evalute(srvsByDomain map[string][]*dns.SRV) (pbuf *bytes.Buffer, err error) {
	input, err := ioutil.ReadFile(tmpl.Src)

	if err != nil {
		return
	}

	vars := map[string]interface{}{
		"domains": srvsByDomain,
	}

	name := filepath.Base(tmpl.Src)
	buf, err := sigil.Execute(input, vars, name)

	if err != nil {
		return
	}

	pbuf = &buf
	return
}

func (tmpl *Template) createTempDest(buf *bytes.Buffer) (tempPath string, err error) {
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

	var destBak string

	if _, e := os.Stat(tmpl.Dest); !os.IsNotExist(e) {
		destBak = tmpl.Dest + ".bak"
		err = Copy(tmpl.Dest, destBak)

		if err != nil {
			return
		}

		defer os.Remove(destBak)
	}

	if tmpl.Dryrun {
		log.Println("*** It does not update the configuration file because it is in dry run mode ***")
		newDest, _ := ioutil.ReadFile(tempPath)
		log.Printf("The new configuration file is as follows:\n---\n%s\n---\n", newDest)
		return
	}

	err = os.Rename(tempPath, tmpl.Dest)

	if err != nil {
		return
	}

	log.Printf("Run '%s' for reloading", tmpl.ReloadCmd.Cmdline)
	err = tmpl.ReloadCmd.Run(tempPath)

	if err != nil {
		err = fmt.Errorf("Reload command failed: %s", err)

		if destBak == "" {
			os.Remove(tmpl.Dest)
		} else {
			os.Rename(destBak, tmpl.Dest)
		}

		return
	}

	return
}

func (tmpl *Template) Process(srvsByDomain map[string][]*dns.SRV) (updated bool) {
	buf, err := tmpl.evalute(srvsByDomain)

	if err != nil {
		tmpl.Status.Ok = false
		log.Println("ERROR: Template evaluating failed:", err)
		return
	}

	tempPath, err := tmpl.createTempDest(buf)

	if err != nil {
		tmpl.Status.Ok = false
		log.Println("ERROR: Temporary dest file creation failed:", err)
		return
	}

	defer os.Remove(tempPath)

	if tmpl.isChanged(tempPath) {
		log.Println("The configuration has been changed. Update", tmpl.Dest)
		err = tmpl.update(tempPath)

		if err != nil {
			tmpl.Status.Ok = false
			log.Println("ERROR: The configuration updating failed:", err)
			return
		}

		updated = true
	}

	tmpl.Status.Ok = true
	return
}
