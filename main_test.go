package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bouk/monkey"
)

func TestMain(t *testing.T) {
	assert := assert.New(t)
	var isParseFlagCalled bool
	var isLoadConfigCalled bool
	var isNewWorkerCalled bool
	var isNewHttpdCalled bool

	monkey.Patch(ParseFlag, func() (_ *Flags) {
		defer monkey.Unpatch(ParseFlag)
		isParseFlagCalled = true
		return
	})

	monkey.Patch(LoadConfig, func(_ *Flags) (_ *Config, _ error) {
		defer monkey.Unpatch(LoadConfig)
		isLoadConfigCalled = true
		return
	})

	monkey.Patch(NewWorker, func(_ *Config, _ chan bool, doneChan chan error, _ chan Status) (worker *Worker) {
		defer monkey.Unpatch(NewWorker)
		worker = &Worker{DoneChan: doneChan}
		isNewWorkerCalled = true

		patchInstanceMethod(worker, "Run", func(guard **monkey.PatchGuard) interface{} {
			return func(w *Worker) {
				defer (*guard).Unpatch()
				(*guard).Restore()
				close(w.DoneChan)
			}
		})

		return
	})

	monkey.Patch(NewHttpd, func(_ *Config, _ chan Status) (httpd *Httpd) {
		defer monkey.Unpatch(NewHttpd)
		httpd = &Httpd{}
		isNewHttpdCalled = true

		patchInstanceMethod(httpd, "Run", func(guard **monkey.PatchGuard) interface{} {
			return func(_ *Httpd) {
				defer (*guard).Unpatch()
				(*guard).Restore()
			}
		})

		return
	})

	main()
	assert.Equal(true, isParseFlagCalled)
	assert.Equal(true, isLoadConfigCalled)
	assert.Equal(true, isNewWorkerCalled)
	assert.Equal(true, isNewHttpdCalled)
}
