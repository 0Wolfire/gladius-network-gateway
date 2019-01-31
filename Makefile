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
all: clean network-gateway

clean:
	@rm -rf ./build/*
	@$(GOMOD) go mod tidy
	@$(GOCLEAN) $(CTL_SRC)

# dependency management
dependencies:
	# download our modules
	@$(GOMOD) go mod download

lint:
	@gometalinter --linter='vet:go tool vet -printfuncs=Infof,Debugf,Warningf,Errorf:PATH:LINE:MESSAGE' cmd/main.go

test: $(CTL_SRC)
	@$(GOTEST) $(CTL_SRC)

network-gateway: test
	@$(GOBUILD) -o $(CTL_DEST)$(BINARY_SUFFIX) $(CTL_SRC)