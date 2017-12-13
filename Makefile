OUTPUT ?= bin
GOOS ?= linux
GOARCH ?= amd64

TIMEOUT ?= 10m

PACKAGE = github.com/zero-os/0-stor
COMMIT_HASH = $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_DATE = $(shell date +%FT%T%z)

SERVER_PACKAGES = $(shell go list ./server/...)
CLIENT_PACKAGES = $(shell go list ./client/...)
BENCH_PACKAGES = $(shell go list ./benchmark/...)

ldflags = -extldflags "-static"
ldflagsversion = -X main.CommitHash=$(COMMIT_HASH) -X main.BuildDate=$(BUILD_DATE) -s -w

all: server cli bench

cli: $(OUTPUT)
ifeq ($(GOOS), darwin)
	GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags '$(ldflagsversion)' -o $(OUTPUT)/zerostorcli ./cmd/zerostorcli
else
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags '$(ldflags)$(ldflagsversion)' -o $(OUTPUT)/zerostorcli ./cmd/zerostorcli
endif

server: $(OUTPUT)
ifeq ($(GOOS), darwin)
	GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags '$(ldflagsversion)' -o $(OUTPUT)/zerostorserver ./cmd/zerostorserver
else
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags '$(ldflags)$(ldflagsversion)' -o $(OUTPUT)/zerostorserver ./cmd/zerostorserver
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
	cp $(OUTPUT)/zerostorcli $(GOPATH)/bin/zerostorcli
	cp $(OUTPUT)/zerostorserver $(GOPATH)/bin/zerostorserver
	cp $(OUTPUT)/zstorbench $(GOPATH)/bin/zstorbench

test: testserver testclient testbench

testrace: testserverrace testclientrace

testserver:
	go test -v -timeout $(TIMEOUT) $(SERVER_PACKAGES)

testclient:
	go test  -v -timeout $(TIMEOUT) $(CLIENT_PACKAGES)

testbench:
	go test  -v $(BENCH_PACKAGES)

testserverrace:
	go test  -v -race $(SERVER_PACKAGES)

testclientrace:
	go test  -v -race $(CLIENT_PACKAGES)

testracebench:
	go test  -v -race $(BENCH_PACKAGES)

testcodegen:
	./utils/scripts/test_codegeneration.sh

update_deps:
	dep ensure -v
	dep ensure -update -v
	dep prune -v
	# ensure we don't delete files that we don't want to prune
	git checkout -- vendor/zombiezen.com/go/capnproto2/std/go.capnp
	git checkout -- vendor/zombiezen.com/go/capnproto2/capnpc-go
	git checkout -- vendor/github.com/golang/protobuf/protoc-gen-go

$(OUTPUT):
	mkdir -p $(OUTPUT)

.PHONY: $(OUTPUT) zerostorcli 0storserver test testserver testclient testserverrace testcodegen testclientrace testbench testracebench
