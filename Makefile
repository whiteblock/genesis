SHELL := /bin/bash

GO111MODULE = on

build:
	go build ./...
.PHONY: build

test:
	go test ./...
.PHONY: test

test_race:
	go test ./... -race -coverprofile=coverage.txt -covermode=atomic
.PHONY: test_race

lint:
	golint ./...
.PHONY: lint

vet:
	go vet ./...
.PHONY: vet

install-dev:
	go get github.com/whiteblock/genesis
.PHONY: install-dev



