# libra
PLATFORM     ?= linux
LDFLAGS      ?= -w -s
REGISTER     ?= yvoilee/pod_tcpstate_exporter
NAME          = pod_tcpstate_exporter
VERSION      ?= 0.0.3

# go
GOPATH       := $(shell go env GOPATH)

PACKAGE       = github.com/yvoilee/pod_tcpstate_exporter
COMMA        := ,
EMPTY        :=
SPACE        := $(EMPTY) $(EMPTY)

# docker build
build: cmd-build
build: docker-build

.PHONE: cmd-build
cmd-build:
	@rm -rf bin/*
	CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ GOARCH=amd64 GOOS=linux CGO_ENABLED=1 \
		go build -ldflags "-X main.buildVersion=$(VERSION) \
		-X main.buildTime=$(shell date +%Y-%m-%d_%H:%M:%S) \
		-X main.buildGitRevision=$(shell git rev-parse HEAD) \
		-X main.buildUser=$(USER) -linkmode external -extldflags -static" -o bin/$(NAME) $(PACKAGE);

.PHONE: docker-build
docker-build:
	docker build -t $(REGISTER):$(VERSION) -f build/Dockerfile --rm .;
	docker push $(REGISTER):$(VERSION);