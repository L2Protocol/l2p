FROM golang:1.27-alpine as l2p

RUN apk add --no-cache make cmake gcc musl-dev linux-headers git bash build-base libc-dev

ADD . /l2p
WORKDIR /l2p
ENV CGO_CFLAGS="-O -D__BLST_PORTABLE__" 
ENV CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"
RUN make geth
RUN mv /l2p/build/bin/geth /usr/local/bin/geth

EXPOSE 8545 8547 31398 31398/udp
ENTRYPOINT [ "/usr/local/bin/geth" ]

FROM ethereum/solc:0.6.4-alpine as l2p-genesis

RUN apk add --no-cache ca-certificates npm nodejs bash alpine-sdk expect

RUN git clone https://github.com/bnb-chain/bsc-genesis-contract.git /root/genesis \
    && cd /root/genesis && npm install
#RUN curl -L https://foundry.paradigm.xyz | bash

COPY --from=l2p /usr/local/bin/geth /usr/local/bin/geth

ENTRYPOINT [ "/bin/bash" ]
