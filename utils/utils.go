package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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
		log.Fatalf("FATAL: %s", err)
	}

	return hex.EncodeToString(h.Sum(nil))
}

func Copy(src string, dest string) (err error) {
	cmd := exec.Command("cp", "-p", src, dest)
	out, err := cmd.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("%s: %s", err, out)
	}

	return
}
