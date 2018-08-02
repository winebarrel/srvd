SHELL   := /bin/bash
PROGRAM := srvd
VERSION := v0.1.0
GOOS    := $(shell go env GOOS)
GOARCH  := $(shell go env GOARCH)
SRC     := $(wildcard **/*.go)

.PHONY: all
all: $(PROGRAM)

.PHONY: dep-ensure
dep-ensure: clean-vendor
	dep ensure

$(PROGRAM): $(SRC)
ifeq ($(GOOS),linux)
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(VERSION)" -a -tags netgo -installsuffix netgo -o pkg/$(PROGRAM)
	[[ "`ldd pkg/$(PROGRAM)`" =~ "not a dynamic executable" ]] || exit 1
else
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(VERSION)" -o pkg/$(PROGRAM)
endif

.PHONY: clean
clean:
	rm -f pkg/*

.PHONY: clean-vendor
clean-vendor:
	rm -rf vendor

.PHONY: package
package: clean $(PROGRAM)
	gzip -c pkg/$(PROGRAM) > pkg/$(PROGRAM)-$(VERSION)-$(GOOS)-$(GOARCH).gz
	rm pkg/$(PROGRAM)

.PHONY: install-dep
install-dep:
	go get -u github.com/golang/dep/cmd/dep

.PHONY: package/linux
package/linux: install-dep dep-ensure
	docker run  -v $(shell pwd):/go/src/github.com/winebarrel/$(PROGRAM) --rm golang \
		make -C /go/src/github.com/winebarrel/$(PROGRAM) package clean-vendor

.PHONY: deb
deb:
	docker run  -v $(shell pwd):/go/src/github.com/winebarrel/$(PROGRAM) --rm golang \
		make -C /go/src/github.com/winebarrel/$(PROGRAM) deb/docker clean-vendor

.PHONY: deb/docker
deb/docker: install-dep dep-ensure
	apt-get update
	apt-get install -y debhelper
	dpkg-buildpackage -us -uc
	mv ../srvd_* pkg/
	rm pkg/$(PROGRAM)
