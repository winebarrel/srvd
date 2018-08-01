SHELL   := /bin/bash
PROGRAM := srvd
VERSION := v0.1.0
SRC     := $(wildcard **/*.go)

.PHONY: all
all: $(PROGRAM)

.PHONY: dep-ensure
dep-ensure:
	dep ensure

$(PROGRAM): $(SRC)
	go build -ldflags "-X main.version=$(VERSION)" -o pkg/$(PROGRAM)

.PHONY: clean
clean:
	rm -f pkg/*
