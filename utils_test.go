package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMd5(t *testing.T) {
	assert := assert.New(t)

	tempFile("hello", func(f *os.File) {
		assert.Equal("5d41402abc4b2a76b9719d911017c592", Md5(f.Name()))
	})
}

func TestCopy(t *testing.T) {
	assert := assert.New(t)

	tempFile("hello", func(f *os.File) {
		dest := f.Name() + "2"
		Copy(f.Name(), dest)
		destContent, _ := ioutil.ReadFile(dest)
		assert.Equal("hello", string(destContent))
	})
}
