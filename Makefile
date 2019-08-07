all: client server

submodules:
	git submodule update --init --remote --jobs 10
	git submodule status

server:
	./build/bin/build-server

client:
	./build/bin/build-client

clean:
	rm -rf tmp results
