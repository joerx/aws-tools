output := $(shell basename $${PWD})

build: vendor
	go build -o $(output)

clean:
	rm $(output)

vendor:
	dep ensure

run: build
	./$(output)
