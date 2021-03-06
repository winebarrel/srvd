SHELL     := /bin/bash
PROGRAM   := srvd
VERSION   := v0.3.8
GOVERSION := 1.12.1
GOOS      := $(shell go env GOOS)
GOARCH    := $(shell go env GOARCH)
TEST_SRC  := $(wildcard test_*.go) $(wildcard *_test.go) $(wildcard */*_test.go)
SRC       := $(filter-out $(TEST_SRC),$(wildcard *.go) $(wildcard */*.go))

.PHONY: all
all: $(PROGRAM)

$(PROGRAM): $(SRC) test
ifeq ($(GOOS),linux)
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(VERSION)" -a -tags netgo -installsuffix netgo -o pkg/$(PROGRAM)
	[[ "`ldd pkg/$(PROGRAM)`" =~ "not a dynamic executable" ]] || exit 1
else
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(VERSION)" -o pkg/$(PROGRAM)
endif

.PHONY: lint
lint:
	golint -set_exit_status . tmplfuncs utils testutils

.PHONY: test
test: $(SRC) $(TEST_SRC) lint
	go test -v -gcflags '-N -l' ./...

.PHONY: clean
clean:
	rm -f pkg/*

.PHONY: package
package: clean $(PROGRAM)
	gzip -c pkg/$(PROGRAM) > pkg/$(PROGRAM)-$(VERSION)-$(GOOS)-$(GOARCH).gz
	rm pkg/$(PROGRAM)

.PHONY: install-golint
install-golint:
	GO111MODULE=off go get -u golang.org/x/lint/golint

.PHONY: package/linux
package/linux:
	docker run -v $(shell pwd):/go/src/github.com/winebarrel/$(PROGRAM) -e GO111MODULE=on --rm golang:$(GOVERSION) \
		make -C /go/src/github.com/winebarrel/$(PROGRAM) \
			install-golint package

.PHONY: deb
deb:
	docker run -v $(shell pwd):/go/src/github.com/winebarrel/$(PROGRAM) -e GO111MODULE=on --rm golang:$(GOVERSION) \
		make -C /go/src/github.com/winebarrel/$(PROGRAM) deb/docker

.PHONY: deb/docker
deb/docker: install-golint
	apt-get update
	apt-get install -y debhelper
	dpkg-buildpackage -us -uc
	mv ../srvd_* pkg/
	rm pkg/$(PROGRAM)
