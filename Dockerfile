FROM golang:1.11.3-stretch as built

ADD . /genesis

#Currently depends on geth, however, this should be removed in the near future
#RUN wget https://gethstore.blob.core.windows.net/builds/geth-linux-amd64-1.7.3-4bb3c89d.tar.gz &&\
RUN wget https://gethstore.blob.core.windows.net/builds/geth-linux-amd64-1.8.20-24d727b6.tar.gz &&\
    tar -xzf geth-linux-amd64-1.8.20-24d727b6.tar.gz &&\
    mv geth-linux-amd64-1.8.20-24d727b6/geth /bin/ &&\
    rm -rf geth-linux-amd64-1.8.20-24d727b6*

RUN cd /genesis &&\
    go get || \
    go build

WORKDIR /genesis


FROM ubuntu:latest as final

RUN mkdir -p /genesis

COPY --from=built /genesis/blockchains /genesis/blockchains
cOPY --from=built /genesis/config.json /genesis/config.json
COPY --from=built /genesis/genesis /genesis/genesis
COPY --from=built /bin/geth /bin/geth



ENTRYPOINT ["/genesis/genesis"]