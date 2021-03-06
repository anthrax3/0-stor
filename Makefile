OUTPUT ?= bin
GOOS ?= linux
GOARCH ?= amd64

TIMEOUT ?= 10m
RACE_TIMEOUT ?= 20m

PACKAGE = github.com/zero-os/0-stor
COMMIT_HASH = $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_DATE = $(shell date +%FT%T%z)

SERVER_PACKAGES = $(shell go list ./server/...)
CLIENT_PACKAGES = $(shell go list ./client/...)
DAEMON_PACKAGES = $(shell go list ./daemon/...)
CMD_PACKAGES = $(shell go list ./cmd/...)
BENCH_PACKAGES = $(shell go list ./benchmark/...)

ldflags = -extldflags "-static"
ldflagsversion = -X $(PACKAGE)/cmd.CommitHash=$(COMMIT_HASH) -X $(PACKAGE)/cmd.BuildDate=$(BUILD_DATE) -s -w

all: client server bench

client: $(OUTPUT)
ifeq ($(GOOS), darwin)
	GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags '$(ldflagsversion)' -o $(OUTPUT)/zstor ./cmd/zstor
else
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags '$(ldflags)$(ldflagsversion)' -o $(OUTPUT)/zstor ./cmd/zstor
endif

server: $(OUTPUT)
ifeq ($(GOOS), darwin)
	GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags '$(ldflagsversion)' -o $(OUTPUT)/zstordb ./cmd/zstordb
else
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags '$(ldflags)$(ldflagsversion)' -o $(OUTPUT)/zstordb ./cmd/zstordb
endif

bench: $(OUTPUT)
ifeq ($(GOOS), darwin)
	GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags '$(ldflagsversion)' -o $(OUTPUT)/zstorbench ./cmd/zstorbench
else
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags '$(ldflags)$(ldflagsversion)' -o $(OUTPUT)/zstorbench ./cmd/zstorbench
endif

install: all
	cp $(OUTPUT)/zstor $(GOPATH)/bin/zstor
	cp $(OUTPUT)/zstordb $(GOPATH)/bin/zstordb
	cp $(OUTPUT)/zstorbench $(GOPATH)/bin/zstorbench

test: testserver testclient testdaemon testcmd testbench

testcov:
	utils/scripts/coverage_test.sh

testrace: testserverrace testclientrace testdaemonrace testbenchrace

testserver:
	go test -v -timeout $(TIMEOUT) $(SERVER_PACKAGES)

testclient:
	go test -v -timeout $(TIMEOUT) $(CLIENT_PACKAGES)

testdaemon:
	go test -v -timeout $(TIMEOUT) $(DAEMON_PACKAGES)

testcmd:
	go test -v -timeout $(TIMEOUT) $(CMD_PACKAGES)

testbench:
	go test -v -timeout $(TIMEOUT) $(BENCH_PACKAGES)

testserverrace:
	go test -race -timeout $(RACE_TIMEOUT) $(SERVER_PACKAGES)

testclientrace:
	go test -race -timeout $(RACE_TIMEOUT) $(CLIENT_PACKAGES)

testdaemonrace:
	go test -v -race -timeout $(RACE_TIMEOUT) $(DAEMON_PACKAGES)

testbenchrace:
	go test -v -race -timeout $(RACE_TIMEOUT) $(BENCH_PACKAGES)

testcodegen:
	./utils/scripts/test_codegeneration.sh

ensure_deps:
	dep ensure -v
	make prune_deps

add_dep:
	dep ensure -v
	dep ensure -v -add $$DEP
	make prune_deps

update_dep:
	dep ensure -v
	dep ensure -v -update $$DEP
	make prune_deps

update_deps:
	dep ensure -v
	dep ensure -update -v
	make prune_deps

prune_deps:
	./utils/scripts/prune_deps_safe.sh

$(OUTPUT):
	mkdir -p $(OUTPUT)

.PHONY: $(OUTPUT) client server install test testcov testrace testserver testclient testdaemon testcmd testbench testserverrace testclientrace testdaemonrace testracebench testcodegen ensure_deps add_dep update_dep update_deps prune_deps
