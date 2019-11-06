#!/bin/sh

set -o errexit
set -o xtrace
set -o pipefail

go get github.com/golang/mock/gomock
go install github.com/golang/mock/mockgen
mockgen -destination=./ssh/mocks/client_mock.go -source=./ssh/client.go -package=mocks
go get -u golang.org/x/lint/golint

if [ -n "$(gofmt -l .)" ]; then
  echo "Go code is not formatted:"
  gofmt -d .
  exit 1
else
  echo "Go code is well formatted"
fi

golint -set_exit_status $(go list ./... | grep -v mocks)
go vet $(go list ./... | grep -v mocks)
go test ./...
go get ./...
go test ./... -coverprofile=coverage.txt -covermode=atomic

chmod 777 coverage.txt 
chmod 777 -R ssh