# Support setting various labels on the final image
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

# Build Geth in a stock Go builder container
FROM golang:1.27-alpine AS builder

RUN apk add --no-cache make cmake gcc musl-dev linux-headers git bash build-base libc-dev
# Get dependencies - will also be cached if we won't change go.mod/go.sum
COPY go.mod /go-ethereum/
COPY go.sum /go-ethereum/
RUN cd /go-ethereum && go mod download

ADD . /go-ethereum

# For blst
ENV CGO_CFLAGS="-O -D__BLST_PORTABLE__" 
ENV CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"
RUN cd /go-ethereum && go run build/ci.go install -static ./cmd/geth

# Pull Geth into a second stage deploy alpine container
FROM alpine:3.21

ARG L2P_USER=l2p
ARG L2P_USER_UID=1000
ARG L2P_USER_GID=1000

ENV L2P_HOME=/l2p
ENV HOME=${L2P_HOME}
ENV DATA_DIR=/data

ENV PACKAGES ca-certificates jq \
  bash bind-tools tini \
  grep curl sed gcc

RUN apk add --no-cache $PACKAGES \
  && rm -rf /var/cache/apk/* \
  && addgroup -g ${L2P_USER_GID} ${L2P_USER} \
  && adduser -u ${L2P_USER_UID} -G ${L2P_USER} --shell /sbin/nologin --no-create-home -D ${L2P_USER} \
  && addgroup ${L2P_USER} tty \
  && sed -i -e "s/bin\/sh/bin\/bash/" /etc/passwd  

RUN echo "[ ! -z \"\$TERM\" -a -r /etc/motd ] && cat /etc/motd" >> /etc/bash/bashrc

WORKDIR ${L2P_HOME}

COPY --from=builder /go-ethereum/build/bin/geth /usr/local/bin/

COPY docker-entrypoint.sh ./

RUN chmod +x docker-entrypoint.sh \
    && mkdir -p ${DATA_DIR} \
    && chown -R ${L2P_USER_UID}:${L2P_USER_GID} ${L2P_HOME} ${DATA_DIR}

VOLUME ${DATA_DIR}

USER ${L2P_USER_UID}:${L2P_USER_GID}

# rpc ws graphql p2p
EXPOSE 8545 8546 8547 31398 31398/udp

ENTRYPOINT ["/sbin/tini", "--", "./docker-entrypoint.sh"]