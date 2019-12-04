GO111MODULE=on

PKG_SOURCES=$(wildcard pkg/*/*.go)
DIRECTORIES=$(wildcard util/*/)  $(sort $(dir $(wildcard pkg/*/*/)))
MOCKS=$(foreach x, $(DIRECTORIES), mocks/$(x))
OUTPUT_DIR=./bin

.PHONY: build test test_race lint vet get mocks clean-mocks manual-mocks
.ONESHELL:

all: prep tester genesis

genesis: | get
	go build

prep:
	@mkdir $(OUTPUT_DIR) 2>> /dev/null | true 

tester:
	go build -o $(OUTPUT_DIR)/tester ./cmd/tester 

test:
	go test ./...

test_race:
	go test ./... -race 

lint:
	golint $(go list ./... | grep -v mocks)

vet:
	go vet $(go list ./... | grep -v mocks)

get:
	go get ./...

clean-mocks:
	rm -rf mocks

mocks: $(MOCKS) manual-mocks
	
$(MOCKS): mocks/% : %
	mockery -output=$@ -dir=$^ -all

manual-mocks: clone-definition mock-definition

clone-definition:
	git clone https://github.com/whiteblock/definition.git mocks/.src/definition || true

mock-definition:
	cd mocks/.src/definition/command/ &&\
	mockery -dir=. -output=../../../../mocks/definition/command/ -all && \
	cd -