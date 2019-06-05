GOC=go
GO111MODULE=on

.PHONY: build test test_race lint vet install-deps coverage

all: genesis

genesis:
	$(GOC) build ./...

test:
	go test ./...

test_race:
	go test ./... -race 

coverage:
	go test ./... -coverprofile=coverage.txt -covermode=atomic

lint:
	golint ./...

vet:
	go vet ./...

install-deps:
	go get ./...
