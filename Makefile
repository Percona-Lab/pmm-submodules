all: client server

submodules:
	git submodule update --init --remote
	git submodule status

server: submodules
	./build/bin/build-server

client: submodules
	./build/bin/build-client

clean:
	rm -rf tmp results
