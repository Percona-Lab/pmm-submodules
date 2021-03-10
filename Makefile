.PHONY: all submodules server client clean purge test help default

default: help

help:                       ## Display this help message.
	@echo "Please use \`make <target>\` where <target> is one of:"
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | \
	awk -F ':.*?## ' 'NF==2 {printf "  %-26s%s\n", $$1, $$2}'

all: client server          ## Build client and server.

submodules:                 ## Update all sumodules .
	git submodule update --init --remote --jobs 10
	git submodule status

server:                     ## Build the server.
	./build/bin/build-server

client:                     ## Build the client.
	./build/bin/build-client

clean:                      ## Clean build results.
	rm -rf tmp results sources/pmm-submodules

purge:                      ## Clean cache and leftovers. Please run this when starting a new feature build.
	git reset --hard && git clean -xdff
	git submodule update
	git submodule foreach 'git reset --hard && git clean -xdff'
