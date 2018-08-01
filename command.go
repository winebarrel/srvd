package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"text/template"
	"time"

	"github.com/mattn/go-shellwords"
)

type Command struct {
	Cmdline string
	Timeout time.Duration
}

func NewCommand(cmdline string, timeout int) (cmd *Command) {
	cmd = &Command{
		Cmdline: cmdline,
		Timeout: time.Second * time.Duration(timeout),
	}

	return
}

func makeCmd(cmdArgs []string) (cmd *exec.Cmd, outReader io.ReadCloser, errReader io.ReadCloser, err error) {
	if len(cmdArgs) > 1 {
		cmd = exec.Command(cmdArgs[0], cmdArgs[1:]...)
	} else {
		cmd = exec.Command(cmdArgs[0])
	}

	outReader, err = cmd.StdoutPipe()

	if err != nil {
		return
	}

	errReader, err = cmd.StderrPipe()

	if err != nil {
		return
	}

	return
}

func tailf(name string, reader io.Reader, wg *sync.WaitGroup) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		log.Printf("%s: %s\n", name, scanner.Text())
	}

	wg.Done()
}

func (command *Command) Run(src string) (err error) {
	tmpl, err := template.New(command.Cmdline).Parse(command.Cmdline)

	if err != nil {
		return
	}

	vars := map[string]interface{}{
		"src": src,
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, vars)

	if err != nil {
		return
	}

	cmdArgs, err := shellwords.Parse(buf.String())
	name := cmdArgs[0]

	cmd, outReader, errReader, err := makeCmd(cmdArgs)

	if err != nil {
		return
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go tailf(name+": stdout", outReader, wg)

	wg.Add(1)
	go tailf(name+": stderr", errReader, wg)

	err = cmd.Start()

	if err != nil {
		return
	}

	var timer *time.Timer
	var timeout bool

	timer = time.AfterFunc(command.Timeout, func() {
		timer.Stop()
		cmd.Process.Kill()
		timeout = true
	})

	err = cmd.Wait()
	timer.Stop()
	wg.Wait()

	if timeout {
		err = fmt.Errorf("%s timed out", name)
		return
	}

	if err != nil {
		return
	}

	return
}
