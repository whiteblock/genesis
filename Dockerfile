FROM golang:1.13.4-alpine as build

ENV GO111MODULE on
WORKDIR /go/src/github.com/whiteblock/genesis

RUN apk add git gcc libc-dev

COPY . .
RUN go get && go build

FROM alpine:3.10 as final

RUN apk add ca-certificates
RUN mkdir -p /etc/whiteblock
RUN mkdir -p /genesis

WORKDIR /genesis

COPY --from=build /go/src/github.com/whiteblock/genesis/config/genesis.yaml /etc/whiteblock/genesis.yaml
COPY --from=build /go/src/github.com/whiteblock/genesis/genesis /genesis/genesis


ENTRYPOINT ["/genesis/genesis"]