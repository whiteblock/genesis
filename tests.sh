#!/bin/sh

set -o errexit
set -o xtrace
set -o pipefail


if [ -n "$(gofmt -l .)" ]; then
  echo "Go code is not formatted:"
  gofmt -d .
  exit 1
else
  echo "Go code is well formatted"
fi

golint -set_exit_status ./...
go vet ./...
make mocks
go get ./...
go test ./...
go test ./... -coverprofile=coverage.txt -covermode=atomic

chmod 777 coverage.txt 