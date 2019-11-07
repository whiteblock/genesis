GOC=go
GO111MODULE=on

PKG_SOURCES=$(wildcard pkg/*/*.go)
MOCKS=$(foreach x, $(PKG_SOURCES), mocks/$(x))


.PHONY: build test test_race lint vet install-deps coverage mocks clean-mocks

all: genesis

genesis: | install-deps
	$(GOC) build ./...

test:
	go test ./...

test_race:
	go test ./... -race 

lint:
	golint $(go list ./... | grep -v mocks)

vet:
	go vet $(go list ./... | grep -v mocks)

install-deps:
	go get ./...

clean-mocks:
	rm -rf mocks

mocks: $(MOCKS)
	
$(MOCKS): mocks/%.go : %.go
	mkdir -p $(dir $@)
	mockgen -destination=$@ -source=$^ -package=mocks
	

#$(foreach dir, $(dir $(wildcard pkg/*/)), $(shell mkdir -p $(dir)/mocks))
#$(foreach f, $(PKG_SOURCES), $(shell mockgen -destination=$(dir $(f))mocks/$(notdir $(f)).go -source=$(f) -package=mocks))

#mockgen -destination=./pkg/usecase/mocks/docker.go -source=./pkg/usecase/docker.go -package=mocks

#install-mock:
#	go get github.com/golang/mock/gomock
#	go install github.com/golang/mock/mockgen

