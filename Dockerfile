FROM golang:1.13.4-stretch as built

ENV GO111MODULE on

ADD . /go/src/github.com/whiteblock/genesis

WORKDIR /go/src/github.com/whiteblock/genesis
RUN go get && go build

FROM ubuntu:latest as final
ENV DEBIAN_FRONTEND noninteractive

RUN mkdir -p /genesis &&\
	apt-get update &&\
	apt-get install --no-install-recommends -y ca-certificates &&\
	apt-get clean &&\
	rm -rf /var/lib/apt/lists/* &&\
	mkdir -p /etc/whiteblock

WORKDIR /genesis

COPY --from=built /go/src/github.com/whiteblock/genesis/config/genesis.yaml /etc/whiteblock/genesis.yaml
COPY --from=built /go/src/github.com/whiteblock/genesis/genesis /genesis/genesis

ENTRYPOINT ["/genesis/genesis"]