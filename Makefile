all: client server

submodules:
	git submodule update --init --remote --jobs 10
	git submodule status

server: submodules
	./build/bin/build-server

client: submodules
	./build/bin/build-client

clean:
	rm -rf tmp results
