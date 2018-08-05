package main

import (
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
