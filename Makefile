GOC=go
GO111MODULE=on

.PHONY: build test test_race lint vet install-deps coverage mocks install-mock

all: genesis

genesis: | install-deps
	$(GOC) build ./...

test:
	go test ./...

test_race:
	go test ./... -race 

lint:
	golint $(go list ./... | grep -v mocks)

format:
	go fmt

vet:
	go vet $(go list ./... | grep -v mocks)

install-deps:
	go get ./...

install-mock:
	go get github.com/golang/mock/gomock
	go install github.com/golang/mock/mockgen

mocks:
	mockgen -destination=./ssh/mocks/client_mock.go -source=./ssh/client.go -package=mocks
