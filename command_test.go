package main

import (
	"bytes"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommand(t *testing.T) {
	assert := assert.New(t)
	cmd := NewCommand("echo {{ .src }}", 3)
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	err := cmd.Run("src")
	assert.Equal(nil, err)
	assert.Regexp(regexp.MustCompile(`echo: stdout: src`), buf.String())
	log.SetOutput(os.Stdout)
}

func TestCommandFailed(t *testing.T) {
	assert := assert.New(t)
	cmd := NewCommand("false", 3)
	err := cmd.Run("")
	assert.Equal("exit status 1", err.Error())
}

func TestCommandTimeout(t *testing.T) {
	assert := assert.New(t)
	cmd := NewCommand("sleep 3", 0)
	err := cmd.Run("")
	assert.Equal("sleep timed out", err.Error())
}
