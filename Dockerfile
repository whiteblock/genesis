FROM golang:1.11.3-stretch

ADD . /genesis

RUN cd /genesis &&\
    go get || \
    go build

WORKDIR /genesis

ENTRYPOINT ["/genesis/genesis"]