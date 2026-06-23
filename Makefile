.PHONY: build test clean install

BINARY_NAME=gitli
VERSION=dev
LDFLAGS=-ldflags "-X github.com/kush/gitli/cmd.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)
	go clean

install: build
	sudo mv $(BINARY_NAME) /usr/local/bin/

run: build
	./$(BINARY_NAME)

dev:
	go run . $(ARGS)

.DEFAULT_GOAL := build