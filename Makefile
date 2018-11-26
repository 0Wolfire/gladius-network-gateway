##
## Makefile to test and build the gladius binaries
##

##
# GLOBAL VARIABLES
##

# if we are running on a windows machine
# we need to append a .exe to the
# compiled binary
BINARY_SUFFIX=
ifeq ($(OS),Windows_NT)
	BINARY_SUFFIX=.exe
endif

ifeq ($(GOOS),windows)
	BINARY_SUFFIX=.exe
endif

# code source and build directories
SRC_DIR=./cmd
DST_DIR=./build

CTL_SRC=$(SRC_DIR)/
CTL_SRC_PROF=$(SRC_DIR)/gladius-network-gateway-profiler

BINARY=gladius-network-gateway$(BINARY_SUFFIX)
CTL_DEST=$(DST_DIR)/$(BINARY)

# commands for go
GOMOD=GO111MODULE=on
GOBUILD=$(GOMOD) go build
GOTEST=$(GOMOD) go test
GOCLEAN=$(GOMOD) go clean

##
# MAKE TARGETS
##

# general make targets
all: 
	make clean
	# make dependencies
	# make lint
	make network-gateway

profile-enabled: network-gateway-profile

clean:
	rm -rf ./build/*
	$(GOMOD) go mod tidy
	$(GOCLEAN) $(CTL_SRC)

# dependency management
dependencies:
	# install go packages
	$(GOMOD) go mod vendor

	# Deal with the ethereum cgo bindings
	$(GOMOD) go get github.com/ethereum/go-ethereum

	# Protobuf generation
	$(GOMOD) go get -u github.com/gogo/protobuf/protoc-gen-gogofaster

	# cp -r \
	# "${GOPATH}/src/github.com/ethereum/go-ethereum/crypto/secp256k1/libsecp256k1" \
	# "vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/"

lint:
	gometalinter --linter='vet:go tool vet -printfuncs=Infof,Debugf,Warningf,Errorf:PATH:LINE:MESSAGE' cmd/main.go

test: $(CTL_SRC)
	$(GOTEST) $(CTL_SRC)

# Made for macOS at the moment
# Install gcc cross compilers for macOS
# `brew install mingw-w64` - windows
# `brew install FiloSottile/musl-cross/musl-cross` - linux
release: clean release-win release-linux release-mac

release-win:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(DST_DIR)/release/windows/$(BINARY).exe $(CTL_SRC)
release-linux:
	CGO_ENABLED=1 CC=x86_64-linux-musl-gcc GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(DST_DIR)/release/linux/$(BINARY) $(CTL_SRC)
release-mac:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(DST_DIR)/release/macos/$(BINARY) $(CTL_SRC)

protobuf:
	protoc -I=. -I=$(GOPATH)/src -I=$(GOPATH)/src/github.com/gogo/protobuf/protobuf --gogofaster_out=\
	Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,\
	Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
	Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,\
	Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
	Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:. \
	./pkg/p2p/peer/messages/*.proto

network-gateway: test
	$(GOBUILD) -o $(CTL_DEST) $(CTL_SRC)

network-gateway-profile: test
	$(GOBUILD) -o $(CTL_DEST) $(CTL_SRC_PROF)
