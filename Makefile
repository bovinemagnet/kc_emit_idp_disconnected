# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOBIN=./bin

# Name of your binary executable
BINARY_NAME=kc_emit_idp_disconnected

all: test build

build:
	$(GOBUILD) -o $(GOBIN)/$(BINARY_NAME) -v

test:
	$(GOTEST) -v ./...

clean:
	rm -rf $(GOBIN)

.PHONY: all build test clean
