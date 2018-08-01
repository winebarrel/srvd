package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"
)

func Md5(path string) string {
	f, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	h := md5.New()
	_, err = io.Copy(h, f)

	if err != nil {
		log.Fatal(err)
	}

	return hex.EncodeToString(h.Sum(nil))
}
