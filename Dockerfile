FROM golang:1.11.3-stretch

ADD . /genesis

#Currently depends on geth, however, this should be removed in the near future
RUN git clone https://github.com/ethereum/go-ethereum.git && \
cd go-ethereum && \
make geth

RUN cd /genesis &&\
    go get || \
    go build

WORKDIR /genesis

ENTRYPOINT ["/genesis/genesis"]