FROM golang:1.12.6-alpine as built

ENV GO111MODULE on

RUN apk add git gcc libc-dev

ADD . /go/src/github.com/whiteblock/genesis

WORKDIR /go/src/github.com/whiteblock/genesis
RUN go get && go build

FROM alpine:latest as final

RUN apk add openssh ca-certificates
RUN mkdir -p /etc/whiteblock
RUN mkdir -p /genesis 

WORKDIR /genesis

COPY --from=built /go/src/github.com/whiteblock/genesis/resources /genesis/resources
COPY --from=built /go/src/github.com/whiteblock/genesis/config/genesis.yaml /etc/whiteblock/genesis.yaml
COPY --from=built /go/src/github.com/whiteblock/genesis/genesis /genesis/genesis

RUN ln -s /genesis/resources/geth/ /genesis/resources/ethereum

ENTRYPOINT ["/genesis/genesis"]