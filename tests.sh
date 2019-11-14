#!/bin/bash

set -o errexit
set -o xtrace

if [ -n "$(gofmt -l .)" ]; then
  echo "Go code is not formatted:"
  gofmt -d .
  exit 1
else
  echo "Go code is well formatted"
fi

golint -set_exit_status $(go list ./... | grep -v mocks)

ls pkg | while read line; do
  mockery -output=mocks/pkg/$line/ -dir=pkg/$line/ -all
done
go vet $(go list ./... | grep -v mocks)
go test ./...
go test ./... -coverprofile=coverage.txt -covermode=atomic
bash <(curl -s https://codecov.io/bash)