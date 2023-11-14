ARG GOPROXY
ARG DOCKER_REGISTRY

### Golang target ###
FROM ${DOCKER_REGISTRY}golang:1.20 AS golang
RUN go version


##### Builder target #####
FROM ${DOCKER_REGISTRY}ubuntu:22.04 AS temporal-builder

# Install dependencies
RUN apt-get -qq update && apt-get -qq install -y curl git make


# Set-up go
COPY --from=golang /usr/local/go/ /usr/local/go/
ENV PATH /usr/local/go/bin:$PATH
ENV GO111MODULE=on

WORKDIR /home/builder

# cache Temporal packages as a docker layer
COPY ./temporal-server/go.mod ./temporal-server/go.sum ./temporal-server/
RUN (cd ./temporal-server && go mod download all)

# cache tctl packages as a docker layer
COPY ./tctl-snap/tctl/go.mod ./tctl-snap/tctl/go.sum ./tctl/
RUN (cd ./tctl && go mod download all)

# cache envtmpl packages as a docker layer
COPY ./envtmpl/go.mod ./envtmpl/
RUN (cd ./envtmpl && go mod download all)

# build
COPY . .
RUN (cd ./temporal-server && CGO_ENABLED=0 make temporal-server)
RUN (cd ./tctl-snap/tctl && make build)
RUN (cd ./envtmpl && go build -o envtmpl .)


##### Temporal server #####
FROM ${DOCKER_REGISTRY}ubuntu:22.04 as temporal-server

# Install dependencies
ENV TZ=Etc/UTC
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get -qq update && apt-get -qq install -y curl git tzdata ca-certificates netcat gettext-base

# Set-up go
COPY --from=golang /usr/local/go/ /usr/local/go/
ENV PATH /usr/local/go/bin:$PATH
ENV GO111MODULE=on

# set up nsswitch.conf for Go's "netgo" implementation
# https://github.com/gliderlabs/docker-alpine/issues/367#issuecomment-424546457
# RUN test ! -e /etc/nsswitch.conf && echo 'hosts: files dns' > /etc/nsswitch.conf

SHELL ["/bin/bash", "-c"]

WORKDIR /etc/temporal

ENV TEMPORAL_HOME /etc/temporal
ENV SERVICES "internal-frontend:history:matching:frontend:worker"

# Membership ports used by the multiple services (frontend, history, matching, worker, internal-frontend)
EXPOSE 6933 6934 6935 6939 6936
# GRPC ports used by the multiple services (frontend, history, matching, worker, internal-frontend)
EXPOSE 7233 7234 7235 7239 7236

# TODO switch WORKDIR to /home/temporal and remove "mkdir" and "chown" calls.
RUN addgroup --gid 1000 temporal
RUN adduser --uid 1000 --gid 1000 --disabled-password temporal
RUN mkdir /etc/temporal/config
RUN chown -R temporal:temporal /etc/temporal/config
USER temporal

# binaries
COPY --from=temporal-builder /home/builder/tctl-snap/tctl/tctl /usr/local/bin
COPY --from=temporal-builder /home/builder/tctl-snap/tctl/tctl-authorization-plugin /usr/local/bin
COPY --from=temporal-builder /home/builder/temporal-server/temporal-server /usr/local/bin
COPY --from=temporal-builder /home/builder/envtmpl/envtmpl /usr/local/bin

# configs
COPY ./temporal-server/config/dynamicconfig /etc/temporal/config/dynamicconfig
COPY ./temporal-server/config/development.yaml /etc/temporal/config/development.yaml

# scripts
COPY ./scripts/entrypoint.sh /etc/temporal/entrypoint.sh
COPY ./scripts/start-temporal.sh /etc/temporal/start-temporal.sh

ENTRYPOINT ["/etc/temporal/entrypoint.sh"]
