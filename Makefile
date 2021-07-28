.PHONY: all server client prepare build clean purge help default

default: help

help:                       ## Display this help message.
	@echo "Please use \`make <target>\` where <target> is one of:"
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | \
	awk -F ':.*?## ' 'NF==2 {printf "  %-26s%s\n", $$1, $$2}'

all: client server          ## Build client and server.

server:                     ## Build the server.
	./build/bin/build-server

client:                     ## Build the client.
	./build/bin/build-client

clean:                      ## Clean build results.
	rm -rf tmp results sources/pmm-submodules

