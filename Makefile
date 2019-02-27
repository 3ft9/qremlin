pkgs = $(shell go list ./... | grep -v /vendor/)

PREFIX  ?= $(shell pwd)
BIN_DIR ?= $(shell pwd)

all: format vet test build

style:
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

test:
	@echo ">> running short tests"
	go test -short $(pkgs)

format:
	@echo ">> formatting code"
	go fmt $(pkgs)

vet:
	@echo ">> vetting code"
	go vet $(pkgs)

build:
	@echo ">> building code"
	cd src && go get ./... && GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../qremlin -a

.PHONY: all style format test vet build
