package testutils

import (
	"io/ioutil"
	"net/http"
	"os"
	"reflect"

	"github.com/bouk/monkey"
)

func TempFile(content string, callback func(f *os.File)) {
	tmpfile, _ := ioutil.TempFile("", "fstaid")
	defer os.Remove(tmpfile.Name())
	tmpfile.WriteString(content)
	tmpfile.Sync()
	tmpfile.Seek(0, 0)
	callback(tmpfile)
}

func ReadResponse(res *http.Response) (string, int) {
	defer res.Body.Close()
	content, _ := ioutil.ReadAll(res.Body)
	return string(content), res.StatusCode
}

func PatchMethod(receiver interface{}, methodName string, replacementf func(**monkey.PatchGuard) interface{}) {
	var guard *monkey.PatchGuard
	replacement := replacementf(&guard)
	guard = monkey.PatchInstanceMethod(
		reflect.TypeOf(receiver), methodName, replacement)
}
