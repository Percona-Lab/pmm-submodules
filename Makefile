all: client server

submodules:
	git submodule update --init --remote

server: submodules
	./build/bin/build-server

client: submodules
	./build/bin/build-client

clean:
	rm -rf tmp results
